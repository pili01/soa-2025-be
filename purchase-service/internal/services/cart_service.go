package services

import (
	"errors"
	"fmt"

	"purchase-service/internal/models"
	"purchase-service/internal/repositories"
)

type CartService struct {
	cartRepo      *repositories.ShoppingCartRepository
	itemRepo      *repositories.OrderItemRepository
}

func NewCartService(cartRepo *repositories.ShoppingCartRepository, itemRepo *repositories.OrderItemRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		itemRepo:    itemRepo,
	}
}

func (s *CartService) GetOrCreateCart(touristID int) (*models.ShoppingCart, error) {
	cart, err := s.cartRepo.GetCartByTouristID(touristID)
	if err != nil {
		return nil, err
	}

	if cart == nil {
		cart, err = s.cartRepo.CreateCart(touristID)
		if err != nil {
			return nil, err
		}
	}

	return cart, nil
}

func (s *CartService) AddToCart(touristID int, request *models.AddToCartRequest) error {
	
	cart, err := s.GetOrCreateCart(touristID)
	if err != nil {
		return err
	}

	existingItems, err := s.itemRepo.GetItemsByCartID(cart.ID)
	if err != nil {
		return err
	}

	for _, item := range existingItems {
		if item.TourID == request.TourID {
			return errors.New("Tour already exists in cart")
		}
	}

	item := &models.OrderItem{
		CartID:   cart.ID,
		TourID:   request.TourID,
		TourName: request.TourName,
		Price:    request.Price,
		Quantity: 1,
	}

	err = s.itemRepo.AddItem(item)
	if err != nil {
		return err
	}

	return s.updateCartTotal(cart.ID)
}

func (s *CartService) RemoveFromCart(touristID int, itemID int) error {
	
	cart, err := s.GetOrCreateCart(touristID)
	if err != nil {
		return err
	}

	
	item, err := s.itemRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}

	if item == nil {
		return errors.New("item not found")
	}

	if item.CartID != cart.ID {
		return errors.New("unauthorized")
	}

	
	err = s.itemRepo.RemoveItem(itemID)
	if err != nil {
		return err
	}

	
	return s.updateCartTotal(cart.ID)
}

func (s *CartService) GetCart(touristID int) (*models.ShoppingCartResponse, error) {
	cart, err := s.GetOrCreateCart(touristID)
	if err != nil {
		return nil, err
	}

	items, err := s.itemRepo.GetItemsByCartID(cart.ID)
	if err != nil {
		return nil, err
	}

	// Dodaj logging da vidimo šta se čita iz baze
	var dbTotal float64
	if cart.TotalPrice.Valid {
		dbTotal = cart.TotalPrice.Float64
	}
	fmt.Printf("GetCart: Cart %d from DB - TotalPrice: %.2f (valid: %v), Items count: %d\n", cart.ID, dbTotal, cart.TotalPrice.Valid, len(items))
	for i, item := range items {
		fmt.Printf("  Item %d: TourID=%d, Price=%.2f, Quantity=%d\n", i+1, item.TourID, item.Price, item.Quantity)
	}

	// Računaj total iz item-a da vidimo da li se poklapa sa onim iz baze
	var calculatedTotal float64
	for _, item := range items {
		calculatedTotal += item.Price
	}
	fmt.Printf("GetCart: Calculated total from items: %.2f, DB total: %.2f (valid: %v)\n", calculatedTotal, dbTotal, cart.TotalPrice.Valid)

	// Koristi calculated total umesto DB total ako DB total nije valid
	var responseTotal float64
	if cart.TotalPrice.Valid {
		responseTotal = cart.TotalPrice.Float64
	} else {
		responseTotal = calculatedTotal
	}

	return &models.ShoppingCartResponse{
		ID:         cart.ID,
		TouristID:  cart.TouristID,
		TotalPrice: responseTotal,
		Items:      items,
		CreatedAt:  cart.CreatedAt,
		UpdatedAt:  cart.UpdatedAt,
	}, nil
}

func (s *CartService) updateCartTotal(cartID int) error {
	items, err := s.itemRepo.GetItemsByCartID(cartID)
	if err != nil {
		return err
	}

	var total float64
	for _, item := range items {
		total += item.Price
	}

	// Dodaj logging da vidimo šta se računa
	fmt.Printf("Updating cart %d total price to: %.2f (items count: %d)\n", cartID, total, len(items))
	for i, item := range items {
		fmt.Printf("  Item %d: TourID=%d, Price=%.2f, Quantity=%d\n", i+1, item.TourID, item.Price, item.Quantity)
	}

	err = s.cartRepo.UpdateCartTotal(cartID, total)
	if err != nil {
		fmt.Printf("Error updating cart total: %v\n", err)
		return err
	}

	fmt.Printf("Successfully updated cart %d total price to: %.2f\n", cartID, total)
	return nil
}


