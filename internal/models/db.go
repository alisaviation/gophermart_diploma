package models

import "time"

type DBUser struct {
	ID          uint           `gorm:"primaryKey"`
	Login       string         `gorm:"unique;not null"`
	Password    string         `gorm:"not null"`
	Current     float64        `gorm:"current"`           // Баланс клиента
	Withdrawn   float64        `gorm:"withdrawn"`         // Сумма списания у клиента
	Orders      []DBOrder      `gorm:"foreignKey:UserID"` // Один пользователь => много заказов
	Withdrawals []DBWithdrawal `gorm:"foreignKey:UserID"` // Один пользователь => много списаний
}

type DBOrder struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null"`
	Number     string    `gorm:"unique;not null"`
	Status     string    `gorm:"not null"`
	Accrual    *float64  `gorm:"accrual"`
	UploadedAt time.Time `gorm:"uploaded_at"`
}

type DBWithdrawal struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null"`
	OrderNumber string    `gorm:"unique;not null"`
	Sum         float64   `gorm:"not null"`
	ProcessedAt time.Time `gorm:"processed_at"`
}
