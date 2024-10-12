package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"gophermart/internal/accrual"
	"gophermart/internal/interfaces"
	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserHandler struct {
	UserService          interfaces.UserServiceInterface
	OrderService         interfaces.OrderServiceInterface
	WithdrawService      interfaces.WithdrawRepositoryInterface
	UserBalanceService   interfaces.UserBalanceRepositoryInterface
	AccrualSystemAddress string
	DBConnectionString   string
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userRepository := uh.UserService.GetUserRepository()

	userID := userRepository.GetUserID(user.Username)

	if userID == repository.DatabaseError {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	if userID >= 0 {
		http.Error(w, "Failed to register user", http.StatusConflict)
		return
	}

	orderRepository := uh.OrderService.GetOrderRepository()
	dbStorage := orderRepository.GetDBStorage()
	err := dbStorage.Init(uh.DBConnectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	err = dbStorage.BeginTransaction()

	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	if user, err = uh.UserService.RegisterUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		_ = dbStorage.Rollback()
		return
	}

	if err = uh.UserBalanceService.CreateUserBalance(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		_ = dbStorage.Rollback()
		return
	}

	if err := dbStorage.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(user)

	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   3600,
	})

	dbStorage.Close()

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds middleware.Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authUser, err := uh.UserService.AuthenticateUser(creds.Username, creds.Password)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(authUser)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   3600,
	})

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(middleware.SecretKey))
}

func (uh *UserHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	//userRepository := uh.UserService.GetUserRepository()
	//
	//userID := userRepository.GetUserID(username)

	//if userID < 0 {
	//	http.Error(w, "пользователь не найден", http.StatusNotFound)
	//	return
	//}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Could not read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	orderNumber := string(body)
	isDigit := isDigits(orderNumber)

	if !isDigit {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if !ValidateNumber(orderNumber) {
		http.Error(w, "Неверный формат номера заказа", http.StatusUnprocessableEntity)
		return
	}

	result, err := uh.OrderService.GetOrderID(orderNumber, userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if result == repository.OrderLoadedByAnotherUser {
		http.Error(w, "Номер заказа уже был загружен другим пользователем", http.StatusConflict)
		return
	} else if result == repository.OrderLoaderByThisUser {
		http.Error(w, "Номер заказа уже был загружен этим пользователем", http.StatusOK)
		return
	}

	err = uh.OrderService.SaveOrder(orderNumber, userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		_, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		orderRepository := uh.OrderService.GetOrderRepository()
		dbStorage := orderRepository.GetDBStorage()

		if err := dbStorage.Init(uh.DBConnectionString); err != nil {
			log.Printf("Error while initializing db connection: %v", err)
			return
		}

		if err := dbStorage.BeginTransaction(); err != nil {
			log.Printf("Failed to begin transaction: %v", err)
			return
		}

		var registerResponse accrual.RegisterResponse
		registerResponse, err := accrual.GetOrderInfo(uh.AccrualSystemAddress, orderNumber)
		if err != nil {
			log.Printf("Failed to get order info: %v", err)
			_ = dbStorage.Rollback()
			return
		}

		if registerResponse.Order != "" {
			if err := uh.OrderService.UpdateOrder(orderNumber, registerResponse.Accrual, registerResponse.Status); err != nil {
				_ = dbStorage.Rollback()
				log.Printf("Failed to update order: %v", err)
				return
			}

			if err := uh.UserBalanceService.UpdateUserBalance(registerResponse.Accrual, userID); err != nil {
				_ = dbStorage.Rollback()
				log.Printf("Failed to update user balance: %v", err)
				return
			}

			// Здесь можно выполнить коммит после успешных операций
			if err := dbStorage.Commit(); err != nil {
				log.Printf("Failed to commit transaction: %v", err)
				return
			}
		} else {
			log.Printf("Order not found in accrual service")
			if err := dbStorage.Rollback(); err != nil {
				log.Printf("Failed to rollback transaction: %v", err)
			}
		}
	}()

	wg.Wait()
}

func isDigits(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

func (uh *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	//userRepository := uh.UserService.GetUserRepository()
	//
	//userID := userRepository.GetUserID(username)
	//
	//if userID < 0 {
	//	http.Error(w, "пользователь не найден", http.StatusNotFound)
	//	return
	//}

	orderData, err := uh.OrderService.GetUserOrders(userID)

	if err != nil {
		if errors.Is(err, repository.ErrNoOrdersFound) {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}

		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(orderData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (uh *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	//userRepository := uh.UserService.GetUserRepository()
	//
	//userID := userRepository.GetUserID(username)
	//
	//if userID < 0 {
	//	http.Error(w, "Пользователь не найден", http.StatusNotFound)
	//	return
	//}

	userBalance, err := uh.UserBalanceService.GetUserBalance(userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(userBalance)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (uh *UserHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	//userRepository := uh.UserService.GetUserRepository()
	//
	//userID := userRepository.GetUserID(username)
	//
	//if userID < 0 {
	//	http.Error(w, "Пользователь не найден", http.StatusNotFound)
	//	return
	//}

	var withdraw models.Withdraw
	err := json.NewDecoder(r.Body).Decode(&withdraw)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isDigit := isDigits(withdraw.Order)

	if !isDigit {
		http.Error(w, "Неверный номер заказа", http.StatusUnprocessableEntity)
		return
	}

	result, err := uh.OrderService.GetOrderID(withdraw.Order, userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if result == 1 {
		http.Error(w, "Номер заказа уже был загружен другим пользователем", http.StatusUnprocessableEntity)
		return
	}

	var code int
	code, err = uh.WithdrawService.Withdraw(userID, withdraw.Order, withdraw.Sum)

	if code == repository.WithdrawTransactionError {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if code == repository.NotEnoughFound {
		http.Error(w, "На счету недостаточно средств", http.StatusPaymentRequired)
		return
	}

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	//userRepository := uh.UserService.GetUserRepository()
	//
	//userID := userRepository.GetUserID(username)
	//
	//if userID < 0 {
	//	http.Error(w, "Пользователь не найден", http.StatusNotFound)
	//	return
	//}

	withdrawalInfo, err := uh.WithdrawService.Withdrawals(userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(withdrawalInfo)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func ValidateNumber(orderNumber string) bool {
	orderNumber = strings.ReplaceAll(orderNumber, " ", "")
	orderNumber = strings.ReplaceAll(orderNumber, "-", "")

	for _, char := range orderNumber {
		if char < '0' || char > '9' {
			return false
		}
	}

	sum := 0
	alternate := false
	for i := len(orderNumber) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(orderNumber[i]))
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
