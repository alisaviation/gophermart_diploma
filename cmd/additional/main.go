package additional

import (
	"time"
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

type Order struct {
	number      int
	status      Status
	accrual     int64
	uploaded_at time.Time
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
