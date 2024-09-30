package models

import "time"

type Wallet struct {
	Wallet_id uint      `json:"wallet_id" gorm:"primary_key"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UserRefer uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserRefer"`
}
