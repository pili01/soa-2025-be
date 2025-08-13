package repository

import (
	"database/sql"
	"fmt"
	"stakeholders-service/internal/models"

	"golang.org/x/crypto/bcrypt"
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

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	query := `SELECT id, username, email, role, name, surname, biography, moto, photo_url, is_blocked FROM users ORDER BY id`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
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
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, role, name, surname, biography, moto, photo_url, is_blocked FROM users WHERE id = $1`

	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
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

func (r *UserRepository) UpdateUserBlockStatus(id int, isBlocked bool) error {
	query := `UPDATE users SET is_blocked = $1 WHERE id = $2`

	result, err := r.DB.Exec(query, isBlocked, id)
	if err != nil {
		return fmt.Errorf("failed to update user block status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) UpdateProfile(user *models.User) error {
	query := `UPDATE users SET name = $1, surname = $2, biography = $3, moto = $4, photo_url = $5 WHERE id = $6`
	_, err := r.DB.Exec(query, user.Name, user.Surname, user.Biography, user.Moto, user.PhotoURL, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}
	return nil
}
