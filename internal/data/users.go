package data

import (
    "time"
    "database/sql"
    "github.com/amari03/test1/internal/validator"
)

type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    Role         string    `json:"role"`
    CreatedAt    time.Time `json:"created_at"`
    LastLoginAt  time.Time `json:"last_login_at"`
}

type UserModel struct {
    DB *sql.DB
}

func ValidateUser(v *validator.Validator, user *User) {
    v.Check(user.Email != "", "email", "must be provided")
    v.Check(validator.Matches(user.Email, validator.EmailRX), "email", "must be a valid email address")
}