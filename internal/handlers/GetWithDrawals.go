package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Repinoid/kurs/internal/models"
	"github.com/Repinoid/kurs/internal/securitate"
)

type WithStruct struct {
	Order        string  `json:"order"`
	Sum          float64 `json:"sum"`
	Processed_at string  `json:"processed_at"`
}

func GetWithDrawals(rwr http.ResponseWriter, req *http.Request) {

	UserID, err := securitate.DataBase.LoginByToken(rwr, req)
	if err != nil {
		return
	}

	db := securitate.DataBase.DB
	order := "select ordernumber as number, amount as sum, processed_at from withdrawn where usercode=$1 order by processed_at ;"

	rows, err := db.Query(context.Background(), order, UserID) //
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("db.Query %+v\n", err)
		return
	}

	ord := WithStruct{}
	orda := []WithStruct{}
	var errScan error
	for rows.Next() {
		var tm time.Time
		errScan = rows.Scan(&ord.Order, &ord.Sum, &tm)
		ord.Processed_at = tm.Format(time.RFC3339)
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
		models.Sugar.Debug("No withdrawals\n")
		return
	}
	rwr.WriteHeader(http.StatusOK)
	//	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
	json.NewEncoder(rwr).Encode(orda)
}
