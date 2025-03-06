package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	add "github.com/Tanya1515/gophermarket/cmd/additional"
)

func (GM *Gophermarket) RegisterUser() http.HandlerFunc {
	registerUser := func(rw http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		var err error
		var user add.User

		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while reading new user credentials: %s", err.Error()), http.StatusBadRequest)
			GM.logger.Errorf("Error while reading user credentials: ", err.Error())
			return
		}

		err = json.Unmarshal(buf.Bytes(), &user)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing new user: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		err = GM.storage.CheckUserLogin(user.Login)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while checking user login: %s", err), http.StatusConflict)
			GM.logger.Errorf("Error while checking user login: ", err)
			return
		}

		err = GM.storage.RegisterNewUser(user)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while regestering new user: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while regestering new user: ", err.Error())
			return
		}

		// create JWT token and add it to cookies

		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte(fmt.Sprintf("New user %s sucessfully registered and authenticated!", user.Login)))

	}
	return http.HandlerFunc(registerUser)
}

func (GM *Gophermarket) AuthentificateUser() http.HandlerFunc {
	authentificateUser := func(rw http.ResponseWriter, r *http.Request) {

		var buf bytes.Buffer
		var err error
		var user add.User

		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while reading user credentials: %s", err.Error()), http.StatusBadRequest)
			GM.logger.Errorf("Error while reading user credentials: ", err.Error())
			return
		}

		err = json.Unmarshal(buf.Bytes(), &user)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing user: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		ok, err := GM.storage.CheckUser(user)
		if (err != nil) && ok {
			http.Error(rw, fmt.Sprintf("Error while processing user with login %s: %s", user.Login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while processing user with login ", user.Login, ": ", err.Error())
			return
		}

		if !ok {
			http.Error(rw, fmt.Sprintf("User %s login/password is incorrect", user.Login), http.StatusUnauthorized)
			GM.logger.Errorf("User ", user.Login, " login/password is incorrect")
			return
		}

		// create JWT token and add it to cookies

		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte(fmt.Sprintf("User %s successfully authentificated!", user.Login)))

	}
	return http.HandlerFunc(authentificateUser)
}

func (GM *Gophermarket) GetOrdersInfobyUser() http.HandlerFunc {
	getOrdersInfobyUser := func(rw http.ResponseWriter, r *http.Request) {
		var user add.User
		user.Login = "ozerova"
		orders := make([]add.Order, 0, 10)

		err := GM.storage.GetAllOrders(&orders, user.Login)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while getting order info for user %s: %s", user.Login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while getting order info for user ", user.Login, ": ", err.Error())
			return
		}

		if orders == nil {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		ordersByte, err := json.Marshal(orders)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling orders to bytes: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while getting order info for user: ", err.Error())
			return
		}

		rw.WriteHeader(http.StatusOK)

		rw.Write(ordersByte)

	}
	return http.HandlerFunc(getOrdersInfobyUser)
}

func (GM *Gophermarket) AddOrdersInfobyUser() http.HandlerFunc {
	addOrdersInfobyUser := func(rw http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		var err error
		var orderNumber int
		var user add.User

		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while reading order number for processing it: %s", err.Error()), http.StatusBadRequest)
			GM.logger.Errorf("Error while reading order number for processing it: ", err.Error())
			return
		}

		err = json.Unmarshal(buf.Bytes(), &orderNumber)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing new order: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		if !add.CheckOrderNumber(orderNumber) {
			http.Error(rw, "Order number is invalid", http.StatusPaymentRequired)
			GM.logger.Errorln("Order number is invalid")
			return
		}
		// need to add function AddNewOrder
		err = GM.storage.AddNewOrder(user, orderNumber)
		if err != nil {
			// need to add new codes
			http.Error(rw, fmt.Sprintf("Error while adding new order to database: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusAccepted)

		rw.Write([]byte(fmt.Sprintf("New order %d is processing", orderNumber)))
	}
	return http.HandlerFunc(addOrdersInfobyUser)
}

func (GM *Gophermarket) GetUserBalance() http.HandlerFunc {
	getUserBalance := func(rw http.ResponseWriter, r *http.Request) {
		// передать сюда пользователя

		var user add.User
		user.Login = "ozerova"

		balance, err := GM.storage.GetUserBalance(user.Login)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while getting user %s balance: %s", user.Login, err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		balanceByte, err := json.Marshal(balance)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error while unmarshalling request body for processing new order: %s", err.Error()), http.StatusInternalServerError)
			GM.logger.Errorf("Error while unmarshalling request body for processing new order: ", err.Error())
			return
		}

		rw.WriteHeader(http.StatusOK)

		rw.Write(balanceByte)

	}
	return http.HandlerFunc(getUserBalance)
}

func (GM *Gophermarket) GetUserWastes() http.HandlerFunc {
	getUserWastes := func(rw http.ResponseWriter, r *http.Request) {
		var user add.User
		user.Login = "ozerova"

	}
	return http.HandlerFunc(getUserWastes)
}

func (GM *Gophermarket) PayByPoints() http.HandlerFunc {
	payByPoints := func(rw http.ResponseWriter, r *http.Request) {

	}
	return http.HandlerFunc(payByPoints)
}
