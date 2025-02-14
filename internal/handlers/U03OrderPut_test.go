package handlers

// Basic imports
import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/Repinoid/kurs/internal/rual"
	"github.com/Repinoid/kurs/internal/securitate"
)

func (suite *TstHandlers) Test03OrderPut() {
	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		testName    string
		urla        string
		userName    string
		orderNum    int
		ContentType string
		TokenSuffix string

		want want
	}{
		{
			testName:    "Right PUT",
			urla:        "/api/user/orders",
			userName:    "us1", // first user
			orderNum:    rual.Luhner(1),
			ContentType: "text/plain",
			want: want{
				code:        http.StatusAccepted, // 202
				response:    `{"status":"StatusAccepted"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">",
		},
		{
			testName:    "Right PUT 222",
			urla:        "/api/user/orders",
			userName:    "us222",
			orderNum:    rual.Luhner(2),
			ContentType: "text/plain",
			want: want{
				code:        http.StatusAccepted, // 202
				response:    `{"status":"StatusAccepted"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">",
		},
		{
			testName:    "Already PUT",
			urla:        "/api/user/orders",
			userName:    "us1", // first user
			orderNum:    rual.Luhner(1),
			ContentType: "text/plain",
			want: want{
				code:        http.StatusOK, // 200
				response:    `{"status":"StatusOK"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">",
		},
		{
			testName:    "Other PUT",
			urla:        "/api/user/orders",
			userName:    "us1", // first user but other's order
			orderNum:    rual.Luhner(2),
			ContentType: "text/plain",
			want: want{
				code:        http.StatusConflict,
				response:    `{"status":"StatusConflict"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">",
		},

		{
			testName:    "Wrong Content Type",
			urla:        "/api/user/orders",
			userName:    "us1",
			orderNum:    rual.Luhner(1),
			ContentType: "application/json", // text/plain should be
			want: want{
				code:        http.StatusBadRequest,
				response:    `{"status":"StatusBadRequest"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">",
		},
		{
			testName:    "Wrong PUT user not exist",
			urla:        "/api/user/orders",
			userName:    "us10",
			orderNum:    rual.Luhner(1),
			ContentType: "text/plain",
			want: want{
				code:        http.StatusUnauthorized,
				response:    `{"status":"StatusUnauthorized"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">",
		},
		{
			testName:    "Wrong TOKEN string",
			urla:        "/api/user/orders",
			userName:    "us1",
			orderNum:    rual.Luhner(1),
			ContentType: "text/plain",
			want: want{
				code:        http.StatusUnauthorized,
				response:    `{"status":"StatusUnauthorized"}`,
				contentType: "application/json",
			},
			TokenSuffix: ">>", // bad string
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
			tokenStr := "Bearer <" + token + tt.TokenSuffix

			request := httptest.NewRequest(http.MethodPost, tt.urla, bytes.NewBufferString(strconv.Itoa(tt.orderNum)))
			w := httptest.NewRecorder()
			request.Header.Set("Content-Type", tt.ContentType)
			request.Header.Set("Authorization", tokenStr)
			PutOrder(w, request)
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
