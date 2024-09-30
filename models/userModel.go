package models

import "time"

type User struct {
	User_id   uint      `json:"user_id" gorm:"primary_key;auto_increment"`
	Name      string    `json:"name" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required,min=8"`
	CreatedAt time.Time `json:"created_at"`
}
