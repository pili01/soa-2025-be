package repositories

import (
	"database/sql"
	"time"

	"purchase-service/db"
	"purchase-service/internal/models"
)

type ShoppingCartRepository struct {
	db *db.Database
}

func NewShoppingCartRepository(db *db.Database) *ShoppingCartRepository {
	return &ShoppingCartRepository{db: db}
}

func (r *ShoppingCartRepository) CreateCart(touristID int) (*models.ShoppingCart, error) {
	query := `
		INSERT INTO shopping_carts (tourist_id, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id, tourist_id, total_price, created_at, updated_at`

	now := time.Now()
	cart := &models.ShoppingCart{}
	err := r.db.DB.QueryRow(query, touristID, now, now).Scan(
		&cart.ID, &cart.TouristID, &cart.TotalPrice, &cart.CreatedAt, &cart.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (r *ShoppingCartRepository) GetCartByTouristID(touristID int) (*models.ShoppingCart, error) {
	query := `
		SELECT id, tourist_id, total_price, created_at, updated_at
		FROM shopping_carts
		WHERE tourist_id = $1`

	cart := &models.ShoppingCart{}
	err := r.db.DB.QueryRow(query, touristID).Scan(
		&cart.ID, &cart.TouristID, &cart.TotalPrice, &cart.CreatedAt, &cart.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return cart, nil
}

func (r *ShoppingCartRepository) UpdateCartTotal(cartID int, totalPrice float64) error {
	query := `
		UPDATE shopping_carts
		SET total_price = $1, updated_at = $2
		WHERE id = $3`

	_, err := r.db.DB.Exec(query, totalPrice, time.Now(), cartID)
	return err
}

func (r *ShoppingCartRepository) DeleteCart(cartID int) error {
	query := `DELETE FROM shopping_carts WHERE id = $1`
	_, err := r.db.DB.Exec(query, cartID)
	return err
}
