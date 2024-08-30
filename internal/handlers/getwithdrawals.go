package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Azcarot/GopherMarketProject/internal/storage"
)

func GetWithdrawals() http.Handler {
	withdraw := func(res http.ResponseWriter, req *http.Request) {
		// var buf bytes.Buffer
		token := req.Header.Get("Authorization")
		claims, ok := storage.VerifyToken(token)
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		var userData storage.UserData
		userData.Login = claims["sub"].(string)
		ok, err := storage.CheckUserExists(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		withdrawals, err := storage.GetWithdrawals(storage.DB, userData)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		result, err := json.Marshal(withdrawals)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(result)
	}
	return http.HandlerFunc(withdraw)
}
