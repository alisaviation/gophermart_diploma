package utils

import (
	"regexp"
	"strconv"
)

// ValidateOrderNumber проверяет номер заказа с помощью алгоритма Луна
func ValidateOrderNumber(number string) bool {
	// Проверяем, что номер состоит только из цифр
	if !regexp.MustCompile(`^\d+$`).MatchString(number) {
		return false
	}

	// Алгоритм Луна
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// ValidateLogin проверяет логин пользователя
func ValidateLogin(login string) bool {
	// Логин должен быть не пустым и не слишком длинным
	return len(login) > 0 && len(login) <= 255
}

// ValidatePassword проверяет пароль пользователя
func ValidatePassword(password string) bool {
	// Пароль должен быть не пустым и не слишком длинным
	return len(password) > 0 && len(password) <= 255
}
