package models

import "gorm.io/gorm"

type DBUser struct {
	gorm.Model
	ID          uint           `gorm:"primaryKey"`
	Login       string         `gorm:"unique;not null"`
	Password    string         `gorm:"not null"`
	Orders      []DBOrder      `gorm:"foreignKey:UserID"` // Один пользователь → много заказов
	Balance     DBBalance      `gorm:"foreignKey:UserID"` // Один пользователь → один баланс
	Withdrawals []DBWithdrawal `gorm:"foreignKey:UserID"` // Один пользователь → много списаний
}

type DBOrder struct {
	gorm.Model
	UserID  uint     `gorm:"not null"`
	Number  int64    `gorm:"unique;not null"`
	Status  string   `gorm:"not null"`
	Accrual *float64 `gorm:""`
}

type DBBalance struct {
	gorm.Model
	UserID   uint    `gorm:"not null"`
	Current  float64 `gorm:"not null"`
	Withdraw float64 `gorm:"not null"`
}

type DBWithdrawal struct {
	gorm.Model
	UserID uint    `gorm:"not null"`
	Order  int64   `gorm:"unique;not null"`
	Sum    float64 `gorm:"not null"`
}
