package handlers

import (
	"context"
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

func PutOrder(rwr http.ResponseWriter, req *http.Request) {

	rwr.Header().Set("Content-Type", "application/json")

	if !strings.Contains(req.Header.Get("Content-Type"), "text/plain") {
		rwr.WriteHeader(http.StatusBadRequest) //400 — неверный формат запроса; не text/plain
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debug("not text/plain \n")
		return
	}
	tokenStr := req.Header.Get("Authorization")
	tokenStr, niceP := strings.CutPrefix(tokenStr, "Bearer <") // обрезаем -- Bearer <token>
	tokenStr, niceS := strings.CutSuffix(tokenStr, ">")

	var tokenID int64
	//	err := DataBase.GetIDByToken(context.Background(), tokenStr, &tokenID)	// получаем ID пользователя по полученному токену

	if (!niceP) || (!niceS) || (securitate.DataBase.GetIDByToken(context.Background(), tokenStr, &tokenID) != nil) { // если неверная строка в Authorization - до GetIDByToken дело не дойдёт
		rwr.WriteHeader(http.StatusUnauthorized)            // 401 — неверная пара логин/пароль;
		fmt.Fprintf(rwr, `{"status":"StatusUnauthorized"}`) // либо токена неверный формат, либо по нему нет юзера в базе
		models.Sugar.Debug("Authorization header\n")
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

	orderStr := string(telo)                            // telo - []byte. В нём только номер заказа
	orderNum, err := strconv.ParseInt(orderStr, 10, 64) //
	if err != nil || (!luhn.Valid(int(orderNum))) {     // если не распарсилось или не по ЛУНУ
		rwr.WriteHeader(http.StatusUnprocessableEntity) // 422 — неверный формат номера заказа;
		fmt.Fprintf(rwr, `{"status":"StatusUnprocessableEntity"}`)
		models.Sugar.Debugf("422 — неверный формат номера заказа; %d\n", orderNum)
		return
	}
	var orderID int64
	err = securitate.DataBase.GetIDByOrder(context.Background(), orderNum, &orderID)
	if err != nil { // если такого номера заказа нет в базе вносим его

		orderStat, statusCode := rual.GetFromAccrual(orderStr)

		//err =  // tokenID)	- ID пользователя по полученному токену
		if statusCode != http.StatusOK ||
			securitate.DataBase.UpLoadOrderByID(context.Background(), tokenID, orderNum, orderStat.Status, orderStat.Accrual) != nil {
			rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
			fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
			models.Sugar.Debug("500 — внутренняя ошибка сервера.\n")
			return
		}
		rwr.WriteHeader(http.StatusAccepted) //202 — новый номер заказа принят в обработку;
		fmt.Fprintf(rwr, `{"status":"StatusAccepted"}`)
		return
	}
	if orderID == tokenID {
		rwr.WriteHeader(http.StatusOK) // 200 — номер заказа уже был загружен ЭТИМ пользователем;
		fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
		models.Sugar.Debug("200 — номер заказа уже был загружен ЭТИМ пользователем;\n")
		return
	}
	rwr.WriteHeader(http.StatusConflict) // 409 — номер заказа уже был загружен другим пользователем;
	fmt.Fprintf(rwr, `{"status":"StatusConflict"}`)
	models.Sugar.Debug("409 — номер заказа уже был загружен другим пользователем;\n")
}
