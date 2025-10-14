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
    LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
    
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

// Get a specific user by ID.
func (m UserModel) Get(id string) (*User, error) {
    query := `
        SELECT id, email, password_hash, role, created_at, last_login_at
        FROM users
        WHERE id = $1`

    var user User
    err := m.DB.QueryRow(query, id).Scan(
        &user.ID,
        &user.Email,
        &user.PasswordHash,
        &user.Role,
        &user.CreatedAt,
        &user.LastLoginAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrRecordNotFound
        }
        return nil, err
    }
    return &user, nil
}

// Update a specific user record.
func (m UserModel) Update(user *User) error {
    query := `
        UPDATE users
        SET email = $1, role = $2
        WHERE id = $3`

    args := []interface{}{
        user.Email,
        user.Role,
        user.ID,
    }

    _, err := m.DB.Exec(query, args...)
    return err
}

// Delete a specific user by ID.
    func (m UserModel) Delete(id string) error {
        query := `
            DELETE FROM users
            WHERE id = $1`

        result, err := m.DB.Exec(query, id)
        if err != nil {
            return err
        }

        rowsAffected, err := result.RowsAffected()
        if err != nil {
            return err
        }

        if rowsAffected == 0 {
            return ErrRecordNotFound
        }

        return nil
    }