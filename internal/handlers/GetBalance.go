package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Repinoid/kurs/internal/models"
	"github.com/Repinoid/kurs/internal/securitate"
)

type BalanceStruct struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func GetBalance(rwr http.ResponseWriter, req *http.Request) {

	UserID, err := securitate.DataBase.LoginByToken(rwr, req)
	if err != nil {
		return
	}

	// tokenStr := req.Header.Get("Authorization")
	// tokenStr, niceP := strings.CutPrefix(tokenStr, "Bearer <") // обрезаем -- Bearer <token>
	// tokenStr, niceS := strings.CutSuffix(tokenStr, ">")

	// //	var UserID int64
	// //	err := DataBase.GetIDByToken(context.Background(), tokenStr, &UserID)	// получаем ID пользователя по полученному токену

	// if (!niceP) || (!niceS) || (securitate.DataBase.GetIDByToken(context.Background(), tokenStr, &UserID) != nil) { // если неверная строка в Authorization - до GetIDByToken дело не дойдёт
	// 	rwr.WriteHeader(http.StatusUnauthorized)            // 401 — неверная пара логин/пароль;
	// 	fmt.Fprintf(rwr, `{"status":"StatusUnauthorized"}`) // либо токена неверный формат, либо по нему нет юзера в базе
	// 	models.Sugar.Debug("Authorization header\n")
	// 	return
	// }

	// order := "SELECT (SELECT SUM(o.accrual) FROM orders O where o.usercode=$1) as current, " +
	// 	"(SELECT COALESCE(SUM(w.amount),0) FROM withdrawn w where w.usercode=$1) as withdrawn;"
	order := "SELECT (SELECT SUM(orders.accrual) FROM orders where orders.usercode=$1), " +
		"(SELECT COALESCE(SUM(withdrawn.amount),0) FROM withdrawn  where withdrawn.usercode=$1) ;"

	row := securitate.DataBase.DB.QueryRow(context.Background(), order, UserID)
	var current, withdr float64
	err = row.Scan(&current, &withdr)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError) // //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("row.Scan %+v\n", err)
		return
	}

	bs := BalanceStruct{Current: current - withdr, Withdrawn: withdr} // текущий счёт - сумма бонусов минус сумма списаний

	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(bs)
}
