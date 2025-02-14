package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Repinoid/kurs/internal/models"
	"github.com/Repinoid/kurs/internal/rual"
	"github.com/Repinoid/kurs/internal/securitate"

	"github.com/theplant/luhn"
)

func Withdraw(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		rwr.WriteHeader(http.StatusBadRequest) //400 — неверный формат запроса; не text/plain
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debug("not text/plain \n")
		return
	}

	UserID, err := securitate.DataBase.LoginByToken(rwr, req)
	if err != nil {
		return
	}

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("io.ReadAll %+v\n", err)
		return
	}
	defer req.Body.Close()

	wdrStruct := struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}{}
	errM := json.Unmarshal([]byte(telo), &wdrStruct)

	orderNum, err := strconv.ParseInt(wdrStruct.Order, 10, 64)     //
	if err != nil || errM != nil || (!luhn.Valid(int(orderNum))) { // если не распарсилось или не по ЛУНУ
		rwr.WriteHeader(http.StatusUnprocessableEntity) // 422 — неверный формат номера заказа;
		fmt.Fprintf(rwr, `{"status":"StatusUnprocessableEntity"}`)
		models.Sugar.Debugf("422 — неверный формат номера заказа; %d\n", orderNum)
		return
	}
	var orderID int64
	err = securitate.DataBase.GetIDByOrder(context.Background(), orderNum, &orderID)
	if err != nil { // если такого номера заказа нет в базе вносим его

		db := securitate.DataBase.DB
		//		ordr := "select SUM(accrual) from orders where usercode=$1;"
		ordr := "SELECT (SELECT SUM(orders.accrual) FROM orders where orders.usercode=$1)- " +
			"(SELECT COALESCE(SUM(withdrawn.amount),0) FROM withdrawn where withdrawn.usercode=$1) ;"
		row := db.QueryRow(context.Background(), ordr, UserID) //
		var accs float64                                       // денег на счету
		err := row.Scan(&accs)
		if err != nil {
			rwr.WriteHeader(http.StatusUnprocessableEntity) // 422 — неверный формат номера заказа;
			fmt.Fprintf(rwr, `{"status":"StatusUnprocessableEntity"}`)
			models.Sugar.Debugf("422 — невернная сумма на списание; %d\n", wdrStruct.Sum)
			return
		}
		if wdrStruct.Sum > accs {
			rwr.WriteHeader(http.StatusPaymentRequired) //402 Payment Required
			fmt.Fprintf(rwr, `{"status":"StatusPaymentRequired"}`)
			models.Sugar.Debug("402 Payment Required\n")
			return
		}
		// -------------------------------------------------------------------------
		ordr = "INSERT INTO withdrawn(userCode, orderNumber, amount) VALUES ($1, $2, $3) ;"
		_, err = db.Exec(context.Background(), ordr, UserID, orderNum, wdrStruct.Sum)
		if err != nil {
			rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
			fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
			models.Sugar.Debug("error insert 2 withdrawn.\n")
			return
		}

		orderStat, statusCode := rual.GetFromAccrual(wdrStruct.Order)
		//err =  // UserID)	- ID пользователя по полученному токену
		if statusCode != http.StatusOK ||
			securitate.DataBase.UpLoadOrderByID(context.Background(), UserID, orderNum, orderStat.Status, orderStat.Accrual) != nil {
			rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
			fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
			models.Sugar.Debug("500 — внутренняя ошибка сервера.\n")
			return
		}
		rwr.WriteHeader(http.StatusOK) //
		fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
		return
	}
	// if orderID == UserID {
	// 	rwr.WriteHeader(http.StatusOK) // 200 — номер заказа уже был загружен ЭТИМ пользователем;
	// 	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
	// 	models.Sugar.Debug("200 — номер заказа уже был загружен ЭТИМ пользователем;\n")
	// 	return
	// }
	rwr.WriteHeader(http.StatusUnprocessableEntity) // 422 — неверный формат номера заказа;
	fmt.Fprintf(rwr, `{"status":"StatusUnprocessableEntity"}`)
	models.Sugar.Debug("422 — неверный формат номера заказа;\n")

}
