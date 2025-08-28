package repository

import (
	"database/sql"
	"fmt"
	"stakeholders-service/internal/models"
)

type PositionRepository struct {
	DB *sql.DB
}

func NewPositionRepository(db *sql.DB) *PositionRepository {
	return &PositionRepository{DB: db}
}

func (p *PositionRepository) CreatePosition(position *models.Position) error {
	positionExists, err := p.PositionExists(position.UserId)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if positionExists {
		return p.UpdatePosition(position)
	}

	query := `INSERT INTO positions (user_id, longitude, latitude) 
			VALUES ($1, $2, $3) RETURNING id`

	err = p.DB.QueryRow(query,
		position.UserId,
		position.Longitude,
		position.Latitude,
	).Scan(&position.ID)

	if err != nil {
		return fmt.Errorf("failed to create position of user: %w", err)
	}

	return nil
}

func (p *PositionRepository) GetPositionByUserID(userId int) (*models.Position, error) {
	var position models.Position

	query := `SELECT id, user_id, longitude, latitude FROM positions WHERE user_id = $1`

	err := p.DB.QueryRow(query, userId).Scan(
		&position.ID,
		&position.UserId,
		&position.Longitude,
		&position.Latitude,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user position not found")
		}
		return nil, fmt.Errorf("failed to get position of user: %w", err)
	}

	return &position, nil
}

func (p *PositionRepository) PositionExists(userId int) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM positions WHERE user_id = $1)`

	err := p.DB.QueryRow(query, userId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check for existing position: %w", err)
	}
	return exists, nil
}

func (p *PositionRepository) UpdatePosition(position *models.Position) error {
	query := `UPDATE positions SET longitude = $1, latitude = $2 WHERE user_id = $3`
	
	_, err := p.DB.Exec(query, position.Longitude, position.Latitude, position.UserId)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}
	return nil
}
