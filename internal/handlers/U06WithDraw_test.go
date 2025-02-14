package handlers

// Basic imports
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/Repinoid/kurs/internal/rual"
	"github.com/Repinoid/kurs/internal/securitate"
)

func (suite *TstHandlers) Test06WithDraw() {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		testName    string
		userName    string
		orderNum    int
		ContentType string
		Sum         float64

		want want
	}{
		{
			testName:    "Right Withdraw",
			userName:    "user01",        // first user
			orderNum:    rual.Luhner(74), // accrual loaded until 99 order
			ContentType: "application/json",
			Sum:         11.11,

			want: want{
				code:        http.StatusOK,
				response:    `{"status":"StatusOK"}`,
				contentType: "application/json",
			},
		},
		{
			testName:    "No money",
			userName:    "user01",         // first user
			orderNum:    rual.Luhner(111), // accrual loaded until 99 order
			ContentType: "application/json",
			Sum:         999999.99,

			want: want{
				code:        http.StatusPaymentRequired,
				response:    `{"status":"StatusPaymentRequired"}`,
				contentType: "application/json",
			},
		},
	}

	ctx := context.Background()
	var err error
	securitate.DataBase, err = securitate.ConnectToDB(ctx)
	if err != nil {
		fmt.Printf("database connection error  %v", err)
		return
	}

	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			var token string
			err = securitate.DataBase.GetToken(ctx, tt.userName, &token)
			tokenStr := "Bearer <" + token + ">"

			OrderStr := strconv.Itoa(tt.orderNum)
			wdr := struct {
				Order string  `json:"order"`
				Sum   float64 `json:"sum"`
			}{Order: OrderStr, Sum: tt.Sum}
			ma, _ := json.Marshal(wdr)
			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(ma))
			w := httptest.NewRecorder()
			request.Header.Set("Content-Type", tt.ContentType)
			request.Header.Set("Authorization", tokenStr)
			Withdraw(w, request)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			suite.Require().NoError(err)
			suite.Assert().Equal(tt.want.code, res.StatusCode)
			suite.Assert().Equal(tt.want.contentType, res.Header.Get("Content-Type"))
			suite.Assert().JSONEq(tt.want.response, string(resBody))

		})
	}
}
