package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Repinoid/kurs/internal/models"
	"github.com/Repinoid/kurs/internal/securitate"
)

func LoginUser(rwr http.ResponseWriter, req *http.Request) {
	if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		rwr.WriteHeader(http.StatusBadRequest) //400 — неверный формат запроса;
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debug("not application/json\n")
		return
	}
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("io.ReadAll %+v\n", err)
		return
	}
	defer req.Body.Close()

	logos := struct {
		UserName string `json:"login"`
		Password string `json:"password"`
	}{}
	err = json.Unmarshal([]byte(telo), &logos)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // 400 — неверный формат запроса;
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debugf("json.Unmarshal %+v err %+v\n", logos, err)
		return
	}

	err = securitate.DataBase.IfUserExists(context.Background(), logos.UserName)
	if err != nil {
		fmt.Printf("User does NOT exist ERR %v\n", err)
		rwr.WriteHeader(http.StatusUnauthorized) // 401 — неверная пара логин/пароль;
		fmt.Fprintf(rwr, `{"status":"StatusUnauthorized"}`)
		return
	}
	err = securitate.DataBase.CheckUserPassword(context.Background(), logos.UserName, logos.Password)
	if err != nil {
		fmt.Printf("Wrong password ERR %v\n", err)
		rwr.WriteHeader(http.StatusUnauthorized) // 401 — неверная пара логин/пароль;
		fmt.Fprintf(rwr, `{"status":"StatusUnauthorized"}`)
		return
	}
	Token, err := securitate.BuildJWTString(logos.UserName, []byte(securitate.SECRET_KEY))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	err = securitate.DataBase.UpdateToken(context.Background(), logos.UserName, Token)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError) //500 — внутренняя ошибка сервера.
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		models.Sugar.Debugf("UpdateToken %+v\n", err)
		return
	}
	tok := struct {
		Token string
		Until time.Time
	}{Token: Token, Until: time.Now().Add(securitate.TOKEN_EXP)}
	rwr.WriteHeader(http.StatusOK) // 200 — пользователь успешно зарегистрирован и аутентифицирован;
	json.NewEncoder(rwr).Encode(tok)
}

// --------------------------------------------------------------------------------------------------
