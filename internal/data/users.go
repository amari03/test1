package data

import (
    "context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

    "github.com/amari03/test1/internal/validator"
    "golang.org/x/crypto/bcrypt"
)

// Create a global variable to represent an anonymous user.
var AnonymousUser = &User{}

type User struct {
    ID          string     `json:"id"`
	Email       string     `json:"email"`
	Password    password   `json:"-"` // Use the custom password type. Changed from PasswordHash
	Role        string     `json:"role"`
	Activated   bool       `json:"activated"`
	Version     int        `json:"-"` // Add the version number.
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
    
}

type UserModel struct {
    DB *sql.DB
}

type password struct{
    plaintext *string
    hash    []byte
}

// IsAnonymous checks if a User instance is the anonymous user.
func (u *User) IsAnonymous() bool {
    return u == AnonymousUser
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	// Validate user's email.
	ValidateEmail(v, user.Email)

	// If the plaintext password is not nil, validate it.
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// This check is important. If a password hash is ever nil, it means there's a
	// logic error in our code, and we should panic.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Insert a new user record into the database.
func (m UserModel) Insert(user *User) error {
	query := `
        INSERT INTO users (email, password_hash, role, activated)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	args := []interface{}{user.Email, user.Password.hash, user.Role, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		// Check for a duplicate email error.
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
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
        &user.Password.hash,
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
        SET email = $1, role = $2, activated = $3, version = version + 1
        WHERE id = $4 AND version = $5
        RETURNING version`

    args := []interface{}{
        user.Email,
        user.Role,
        user.Activated,
        user.ID,
        user.Version,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
    if err != nil {
        if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
            return ErrDuplicateEmail
        }
        if errors.Is(err, sql.ErrNoRows) {
            return ErrEditConflict
        }
        return err
    }
    return nil
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

// GetAll returns a slice of all users.
func (m UserModel) GetAll() ([]*User, error) {
    query := `
        SELECT id, email, role, created_at, last_login_at
        FROM users
        ORDER BY email`

    rows, err := m.DB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User

    for rows.Next() {
        var user User
        err := rows.Scan(
            &user.ID,
            &user.Email,
            &user.Role,
            &user.CreatedAt,
            &user.LastLoginAt,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, &user)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}

// Set calculates the bcrypt hash of a plaintext password.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// Matches checks if the provided plaintext password matches the stored hash.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetForToken returns a user record based on a token and its scope.
func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
    tokenHash := sha256.Sum256([]byte(tokenPlaintext))

    query := `
        SELECT users.id, users.email, users.password_hash, users.role, users.activated, users.version, users.created_at
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2
        AND tokens.expiry > $3`

    args := []any{tokenHash[:], tokenScope, time.Now()}

    var user User
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    err := m.DB.QueryRowContext(ctx, query, args...).Scan(
        &user.ID,
        &user.Email,
        &user.Password.hash, // Scan into the hash field of the password struct
        &user.Role,
        &user.Activated,
        &user.Version,
        &user.CreatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrRecordNotFound
        }
        return nil, err
    }

    return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
        SELECT id, email, password_hash, role, activated, version, created_at, last_login_at
        FROM users
        WHERE email = $1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Activated,
		&user.Version,
		&user.CreatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

// UpdatePassword updates the password for a specific user.
func (m UserModel) UpdatePassword(user *User) error {
	query := `
        UPDATE users
        SET password_hash = $1, version = version + 1
        WHERE id = $2 AND version = $3
        RETURNING version`

	args := []any{
		user.Password.hash,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}