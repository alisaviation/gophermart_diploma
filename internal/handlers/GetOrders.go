package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Repinoid/kurs/internal/models"
	"github.com/Repinoid/kurs/internal/securitate"
)

type OrdStruct struct {
	Number      string  `json:"number"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
	Uploaded_at string  `json:"uploaded_at"`
}

func GetOrders(rwr http.ResponseWriter, req *http.Request) {

	tokenStr := req.Header.Get("Authorization")
	tokenStr, niceP := strings.CutPrefix(tokenStr, "Bearer <") // обрезаем -- Bearer <token>
	tokenStr, niceS := strings.CutSuffix(tokenStr, ">")

	var UserID int64
	//	err := DataBase.GetIDByToken(context.Background(), tokenStr, &UserID)	// получаем ID пользователя по полученному токену

	if (!niceP) || (!niceS) || (securitate.DataBase.GetIDByToken(context.Background(), tokenStr, &UserID) != nil) { // если неверная строка в Authorization - до GetIDByToken дело не дойдёт
		rwr.WriteHeader(http.StatusUnauthorized)            // 401 — неверная пара логин/пароль;
		fmt.Fprintf(rwr, `{"status":"StatusUnauthorized"}`) // либо токена неверный формат, либо по нему нет юзера в базе
		models.Sugar.Debug("Authorization header\n")
		return
	}

	db := securitate.DataBase.DB
	order := "select ordernumber as number, orderstatus as status, accrual, uploaded_at from orders where usercode=$1 order by uploaded_at ;"

	rows, err := db.Query(context.Background(), order, UserID) //
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("db.Query %+v\n", err)
		return
	}

	ord := OrdStruct{}
	orda := []OrdStruct{}
	var errScan error
	for rows.Next() {
		var tm time.Time
		errScan = rows.Scan(&ord.Number, &ord.Status, &ord.Accrual, &tm)
		ord.Uploaded_at = tm.Format(time.RFC3339)
		if errScan != nil {
			break
		}
		orda = append(orda, ord)
	}
	rows.Close()
	if err := rows.Err(); err != nil || errScan != nil { // Err returns any error that occurred while reading. Err must only be called after the Rows is closed
		rwr.WriteHeader(http.StatusInternalServerError) // //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("db.Query %+v\n", err)
		return
	}
	if len(orda) == 0 {
		rwr.WriteHeader(http.StatusNoContent) // 204 No Content — сервер успешно обработал запрос, но в ответе были переданы только заголовки без тела сообщения
		fmt.Fprintf(rwr, `{"status":"StatusNoContent"}`)
		models.Sugar.Debug("No ORDERS\n")
		return
	}
	rwr.WriteHeader(http.StatusOK)
	//	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
	json.NewEncoder(rwr).Encode(orda)
}
