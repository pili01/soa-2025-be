package repository

import (
	"fmt"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"stakeholders-service/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	usernameExists, emailExists, err := r.UserExists(user.Username, user.Email)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if usernameExists {
		return fmt.Errorf("username already exists")
	}

	if emailExists {
		return fmt.Errorf("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (username, password, email, role, name, surname, biography, moto, photo_url, is_blocked)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	err = r.DB.QueryRow(query,
		user.Username,
		hashedPassword,
		user.Email,
		user.Role,
		user.Name,
		user.Surname,
		user.Biography,
		user.Moto,
		user.PhotoURL,
		user.IsBlocked,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password, email, role, name, surname, biography, moto, photo_url, is_blocked FROM users WHERE username = $1`

	err := r.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&user.Name,
		&user.Surname,
		&user.Biography,
		&user.Moto,
		&user.PhotoURL,
		&user.IsBlocked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UserExists(username, email string) (bool, bool, error) {
	var usernameExists bool
	var emailExists bool

	usernameQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.DB.QueryRow(usernameQuery, username).Scan(&usernameExists)
	if err != nil {
		return false, false, fmt.Errorf("failed to check for existing username: %w", err)
	}

	emailQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err = r.DB.QueryRow(emailQuery, email).Scan(&emailExists)
	if err != nil {
		return false, false, fmt.Errorf("failed to check for existing email: %w", err)
	}

	return usernameExists, emailExists, nil
}