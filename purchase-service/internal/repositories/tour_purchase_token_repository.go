package repositories

import (
	"database/sql"
	"time"

	"purchase-service/db"
	"purchase-service/internal/models"
)

type TourPurchaseTokenRepository struct {
	db *db.Database
}

func NewTourPurchaseTokenRepository(db *db.Database) *TourPurchaseTokenRepository {
	return &TourPurchaseTokenRepository{db: db}
}

func (r *TourPurchaseTokenRepository) CreateToken(token *models.TourPurchaseToken) error {
	query := `
		INSERT INTO tour_purchase_tokens (tourist_id, tour_id, token, purchased_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	return r.db.DB.QueryRow(query, token.TouristID, token.TourID, token.Token, time.Now()).Scan(&token.ID)
}

func (r *TourPurchaseTokenRepository) GetTokensByTouristID(touristID int) ([]models.TourPurchaseToken, error) {
	query := `
		SELECT id, tourist_id, tour_id, token, purchased_at
		FROM tour_purchase_tokens
		WHERE tourist_id = $1
		ORDER BY purchased_at DESC`

	rows, err := r.db.DB.Query(query, touristID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.TourPurchaseToken
	for rows.Next() {
		var token models.TourPurchaseToken
		err := rows.Scan(&token.ID, &token.TouristID, &token.TourID, &token.Token, &token.PurchasedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *TourPurchaseTokenRepository) GetTokenByValue(tokenValue string) (*models.TourPurchaseToken, error) {
	query := `
		SELECT id, tourist_id, tour_id, token, purchased_at
		FROM tour_purchase_tokens
		WHERE token = $1`

	token := &models.TourPurchaseToken{}
	err := r.db.DB.QueryRow(query, tokenValue).Scan(
		&token.ID, &token.TouristID, &token.TourID, &token.Token, &token.PurchasedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return token, nil
}

func (r *TourPurchaseTokenRepository) ValidateToken(tokenValue string, tourID int) (bool, error) {
	query := `
		SELECT COUNT(*) FROM tour_purchase_tokens
		WHERE token = $1 AND tour_id = $2`

	var count int
	err := r.db.DB.QueryRow(query, tokenValue, tourID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
