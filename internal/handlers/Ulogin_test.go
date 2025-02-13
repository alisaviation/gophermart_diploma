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
	"time"

	"github.com/Repinoid/kurs/internal/rual"
	"github.com/Repinoid/kurs/internal/securitate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *TstHandlers) Test02UserLogin() {
	type logos struct {
		UserName string `json:"login"`
		Password string `json:"password"`
	}
	type want struct {
		code        int
		contentType string
		noMarshErr  bool
	}
	tests := []struct {
		testName string
		urla     string
		userName string
		password string

		want want
	}{
		{
			testName: "Right case",
			urla:     "/api/user/register",
			userName: "us1",
			password: "pass1",
			want: want{
				code:        http.StatusOK,
				noMarshErr:  true,
				contentType: "application/json",
			},
		},
		{
			testName: "Wrong user",
			urla:     "/api/user/register",
			userName: "us11",
			password: "pass1",
			want: want{
				code:        http.StatusUnauthorized, // 401 — неверная пара логин/пароль;
				noMarshErr:  false,
				contentType: "application/json",
			},
		},
		{
			testName: "Wrong password",
			urla:     "/api/user/register",
			userName: "us1",
			password: "pass2",
			want: want{
				code:        http.StatusUnauthorized, // 401 — неверная пара логин/пароль;
				noMarshErr:  false,
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
	defer securitate.DataBase.DB.Close(ctx)

	for _, tt := range tests {
		suite.Run(tt.testName, func() {
			lo, _ := json.Marshal(logos{UserName: tt.userName, Password: tt.password})
			request := httptest.NewRequest(http.MethodPost, tt.urla, bytes.NewBuffer(lo))
			w := httptest.NewRecorder()
			request.Header.Set("Content-Type", "application/json")
			LoginUser(w, request)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.want.code, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				tok := struct {
					Token string
					Until time.Time
				}{}
				err = json.Unmarshal([]byte(resBody), &tok)
				require.NoError(suite.T(), err)
				require.NotEqual(suite.T(), tok.Token, "")

				//	assert.JSONEq(t, tt.want.response, string(resBody))
				suite.Assert().Equal(tt.want.contentType, res.Header.Get("Content-Type"))

				suite.Run(tt.testName, func() { // проверка на вход с токеном, - размещение заказа
					request := httptest.NewRequest(http.MethodPost, "/api/user/orders",
						bytes.NewBufferString(strconv.Itoa(rual.Luhner(4))))
					w := httptest.NewRecorder()
					request.Header.Set("Content-Type", "text/plain")
					request.Header.Set("Authorization", "Bearer <"+tok.Token+">")
					PutOrder(w, request)
					res := w.Result()
					defer res.Body.Close()
					resBody, err := io.ReadAll(res.Body)
					suite.Require().NoError(err)
					suite.Assert().Equal(http.StatusAccepted, res.StatusCode)
					suite.Assert().Equal("application/json", res.Header.Get("Content-Type"))
					suite.Assert().JSONEq(`{"status":"StatusAccepted"}`, string(resBody))

				})
			}
		})
	}
}
