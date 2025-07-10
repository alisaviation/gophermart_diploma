package utils

import "testing"

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{"valid number", "12345678903", true},
		{"valid number 2", "9278923470", true},
		{"invalid number", "1234567890", false},
		{"empty string", "", false},
		{"non-numeric", "123abc456", false},
		{"single digit", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOrderNumber(tt.number)
			if result != tt.expected {
				t.Errorf("ValidateOrderNumber(%s) = %v, want %v", tt.number, result, tt.expected)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		expected bool
	}{
		{"valid login", "user123", true},
		{"empty login", "", false},
		{"very long login", string(make([]byte, 256)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLogin(tt.login)
			if result != tt.expected {
				t.Errorf("ValidateLogin(%s) = %v, want %v", tt.login, result, tt.expected)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"valid password", "password123", true},
		{"empty password", "", false},
		{"very long password", string(make([]byte, 256)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.password)
			if result != tt.expected {
				t.Errorf("ValidatePassword(%s) = %v, want %v", tt.password, result, tt.expected)
			}
		})
	}
}
