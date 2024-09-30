package models

import "time"

type Transaction struct {
	Transaction_id        uint      `json:"transaction_id" gorm:"primary_key;auto_increment"`
	Source_wallet_id      uint      `json:"source_wallet_id"`
	Destination_wallet_id uint      `json:"destination_wallet_id"`
	Amount                float64   `json:"amount"`
	Timestamp             time.Time `json:"Timestamp"`
}
