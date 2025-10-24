package data

import (
	//"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/amari03/test1/internal/validator"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// setupUsersTestDB sets up a clean DB with users and tokens tables.
func setupUsersTestDB(t *testing.T) *sql.DB {
	dsn := "postgres://test1_test:fishsticks@localhost/test1_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err, "failed to connect to the test database")

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	require.NoError(t, err)

	// Create users table
	createUsersTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        email TEXT UNIQUE NOT NULL,
        password_hash BYTEA NOT NULL,
        activated BOOL NOT NULL,
        role TEXT NOT NULL,
        version INTEGER NOT NULL DEFAULT 1,
        last_login_at TIMESTAMPTZ
    );`
	_, err = db.Exec(createUsersTableSQL)
	require.NoError(t, err)

	// Create tokens table for GetForToken test
	createTokensTableSQL := `
    CREATE TABLE IF NOT EXISTS tokens (
        hash BYTEA PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        expiry TIMESTAMPTZ NOT NULL,
        scope TEXT NOT NULL
    );`
	_, err = db.Exec(createTokensTableSQL)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS tokens;")
		require.NoError(t, err)
		_, err = db.Exec("DROP TABLE IF EXISTS users;")
		require.NoError(t, err)
		db.Close()
	})

	return db
}

// newTestUser is a helper to create a valid user instance.
func newTestUser(t *testing.T) *User {
	user := &User{
		Email:     "john.doe@example.com",
		Role:      "user",
		Activated: true,
	}
	err := user.Password.Set("password123")
	require.NoError(t, err)
	return user
}

func TestPassword_SetAndMatches(t *testing.T) {
	var p password
	plaintext := "mysecretpassword"

	err := p.Set(plaintext)
	require.NoError(t, err)

	// Test that the plaintext and hash fields are set
	require.NotNil(t, p.plaintext)
	require.Equal(t, plaintext, *p.plaintext)
	require.NotEmpty(t, p.hash)

	// Test a successful match
	match, err := p.Matches(plaintext)
	require.NoError(t, err)
	require.True(t, match)

	// Test an unsuccessful match
	match, err = p.Matches("wrongpassword")
	require.NoError(t, err)
	require.False(t, match)
}

func TestValidateUser(t *testing.T) {
	// Test case for invalid user data
	v := validator.New()
	user := &User{
		Email: "not-an-email",
		Role:  "user",
	}
	user.Password.plaintext = new(string)
	*user.Password.plaintext = "short"
	user.Password.hash = []byte("dummyhash") // to prevent panic

	ValidateUser(v, user)
	require.False(t, v.Valid())
	require.Contains(t, v.Errors, "email")
	require.Contains(t, v.Errors, "password")

	// Test case for valid user
	v = validator.New()
	validUser := newTestUser(t)
	ValidateUser(v, validUser)
	require.True(t, v.Valid())
}

func TestUserModel_InsertAndGetByEmail(t *testing.T) {
	db := setupUsersTestDB(t)
	m := UserModel{DB: db}
	user := newTestUser(t)

	err := m.Insert(user)
	require.NoError(t, err)

	// Verify ID, CreatedAt, and Version are set
	require.NotEmpty(t, user.ID)
	require.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	require.Equal(t, 1, user.Version)

	// Get by email to confirm
	fetchedUser, err := m.GetByEmail(user.Email)
	require.NoError(t, err)
	require.NotNil(t, fetchedUser)
	require.Equal(t, user.ID, fetchedUser.ID)
	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.Role, fetchedUser.Role)

	// Test duplicate email
	err = m.Insert(user)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDuplicateEmail))
}

func TestUserModel_Update(t *testing.T) {
	db := setupUsersTestDB(t)
	m := UserModel{DB: db}

	user := newTestUser(t)
	err := m.Insert(user)
	require.NoError(t, err)

	// Update fields
	user.Email = "john.doe.updated@example.com"
	user.Role = "admin"
	user.Activated = false

	err = m.Update(user)
	require.NoError(t, err)
	require.Equal(t, 2, user.Version)

	// Fetch to verify
	fetched, err := m.GetByEmail(user.Email)
	require.NoError(t, err)
	require.Equal(t, "admin", fetched.Role)
	require.False(t, fetched.Activated)
	require.Equal(t, 2, fetched.Version)

	// Test edit conflict
	user.Version = 1
	err = m.Update(user)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrEditConflict))

	// Test duplicate email on update
	user2 := &User{Email: "jane.doe@example.com", Role: "user", Activated: true}
	err = user2.Password.Set("password123")
	require.NoError(t, err)
	err = m.Insert(user2)
	require.NoError(t, err)

	fetched.Email = user2.Email // try to update fetched user's email to user2's email
	err = m.Update(fetched)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDuplicateEmail))
}

func TestUserModel_UpdatePassword(t *testing.T) {
	db := setupUsersTestDB(t)
	m := UserModel{DB: db}
	user := newTestUser(t)
	err := m.Insert(user)
	require.NoError(t, err)

	// Set a new password
	newPassword := "new-secure-password-456"
	err = user.Password.Set(newPassword)
	require.NoError(t, err)

	// Update in the database
	err = m.UpdatePassword(user)
	require.NoError(t, err)
	require.Equal(t, 2, user.Version) // Version should increment

	// Fetch and verify the new password works
	fetched, err := m.GetByEmail(user.Email)
	require.NoError(t, err)

	match, err := fetched.Password.Matches(newPassword)
	require.NoError(t, err)
	require.True(t, match)

	// Verify old password fails
	match, err = fetched.Password.Matches("password123")
	require.NoError(t, err)
	require.False(t, match)
}

func TestUserModel_GetForToken(t *testing.T) {
	db := setupUsersTestDB(t)
	m := UserModel{DB: db}

	// 1. Create a user
	user := newTestUser(t)
	err := m.Insert(user)
	require.NoError(t, err)

	// 2. Create a token
	tokenScope := "authentication"
	tokenPlaintext := "supersecrettoken"
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	expiry := time.Now().Add(24 * time.Hour)

	// 3. Insert the token into the DB
	query := "INSERT INTO tokens (hash, user_id, scope, expiry) VALUES ($1, $2, $3, $4)"
	_, err = db.Exec(query, tokenHash[:], user.ID, tokenScope, expiry)
	require.NoError(t, err)

	// 4. Test successful retrieval
	fetchedUser, err := m.GetForToken(tokenScope, tokenPlaintext)
	require.NoError(t, err)
	require.NotNil(t, fetchedUser)
	require.Equal(t, user.ID, fetchedUser.ID)

	// 5. Test sad paths
	// Wrong scope
	_, err = m.GetForToken("wrong-scope", tokenPlaintext)
	require.ErrorIs(t, err, ErrRecordNotFound)
	// Wrong token
	_, err = m.GetForToken(tokenScope, "wrong-token")
	require.ErrorIs(t, err, ErrRecordNotFound)
	// Expired token
	expiredExpiry := time.Now().Add(-1 * time.Hour)
	_, err = db.Exec("UPDATE tokens SET expiry = $1 WHERE hash = $2", expiredExpiry, tokenHash[:])
	require.NoError(t, err)
	_, err = m.GetForToken(tokenScope, tokenPlaintext)
	require.ErrorIs(t, err, ErrRecordNotFound)
}

func TestUserModel_GetAll(t *testing.T) {
	db := setupUsersTestDB(t)
	m := UserModel{DB: db}

	// Insert multiple users
	user1 := newTestUser(t)
	user2 := newTestUser(t)
	user2.Email = "jane.doe@example.com"
	require.NoError(t, m.Insert(user1))
	require.NoError(t, m.Insert(user2))

	users, err := m.GetAll()
	require.NoError(t, err)
	require.Len(t, users, 2)
	// The query orders by email, so jane should be first.
	require.Equal(t, "jane.doe@example.com", users[0].Email)
	require.Equal(t, "john.doe@example.com", users[1].Email)
}
