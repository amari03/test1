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

// Insert a new user record into the database.
func (m UserModel) Insert(user *User) error {
    query := `
        INSERT INTO users (email, password_hash, role)
        VALUES ($1, $2, $3)
        RETURNING id, created_at`

    // We are not hashing yet, just storing the provided string.
    args := []interface{}{user.Email, user.PasswordHash, user.Role}
    return m.DB.QueryRow(query, args...).Scan(&user.ID, &user.CreatedAt)
}