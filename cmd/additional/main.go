package additional

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Status string

const (
	New        Status = "NEW"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)

type OrderSpend struct {
	Number       int       `json:"order"`
	Sum          float64   `json:"sum"`
	Processed_at time.Time `json:"processed_at"`
}

type Order struct {
	Number      int       `json:"number"`
	Status      Status    `json:"status"`
	Accrual     float64   `json:"accrual"`
	Uploaded_at time.Time `json:"uploaded_at"`
}

const TOKEN_EXP = time.Hour

type Claims struct {
	jwt.RegisteredClaims
	UserLogin    string
	UserPassword string
}

func GenerateToken(user User, secretKey string) (JWTtoken string, err error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserLogin:    user.Login,
		UserPassword: user.Password,
	})

	JWTtoken, err = token.SignedString([]byte(secretKey))
	if err != nil {
		return
	}

	return
}

func CheckOrderNumber(orderNumber int) bool {
	var num, sum int
	arrayDigits := make([]int, 0, 10)

	for orderNumber > 0 {
		num = orderNumber % 10
		orderNumber = orderNumber / 10
		arrayDigits = append(arrayDigits, num)
	}

	for key, value := range arrayDigits {
		if ((key + 1) % 2) == ((len(arrayDigits)) % 2) {
			value = value * 2
			if value > 9 {
				value = value - 9
			}
		}
		sum += value
	}

	if sum%10 == 0 {
		return true
	} else {
		return false
	}

}
