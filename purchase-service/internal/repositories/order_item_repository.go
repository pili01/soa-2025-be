package repositories

import (
	"database/sql"
	"time"

	"purchase-service/db"
	"purchase-service/internal/models"
)

type OrderItemRepository struct {
	db *db.Database
}

func NewOrderItemRepository(db *db.Database) *OrderItemRepository {
	return &OrderItemRepository{db: db}
}

func (r *OrderItemRepository) AddItem(item *models.OrderItem) error {
	query := `
		INSERT INTO order_items (cart_id, tour_id, tour_name, price, quantity, added_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	return r.db.DB.QueryRow(query, item.CartID, item.TourID, item.TourName, item.Price, item.Quantity, time.Now()).Scan(&item.ID)
}

func (r *OrderItemRepository) GetItemsByCartID(cartID int) ([]models.OrderItem, error) {
	query := `
		SELECT id, cart_id, tour_id, tour_name, price, quantity, added_at
		FROM order_items
		WHERE cart_id = $1
		ORDER BY added_at DESC`

	rows, err := r.db.DB.Query(query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.CartID, &item.TourID, &item.TourName, &item.Price, &item.Quantity, &item.AddedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *OrderItemRepository) RemoveItem(itemID int) error {
	query := `DELETE FROM order_items WHERE id = $1`
	_, err := r.db.DB.Exec(query, itemID)
	return err
}

func (r *OrderItemRepository) GetItemByID(itemID int) (*models.OrderItem, error) {
	query := `
		SELECT id, cart_id, tour_id, tour_name, price, quantity, added_at
		FROM order_items
		WHERE id = $1`

	item := &models.OrderItem{}
	err := r.db.DB.QueryRow(query, itemID).Scan(
		&item.ID, &item.CartID, &item.TourID, &item.TourName, &item.Price, &item.Quantity, &item.AddedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return item, nil
}

func (r *OrderItemRepository) ClearCart(cartID int) error {
	query := `DELETE FROM order_items WHERE cart_id = $1`
	_, err := r.db.DB.Exec(query, cartID)
	return err
}
