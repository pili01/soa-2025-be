package services

import (
	"errors"

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

	item := &models.OrderItem{
		CartID:   cart.ID,
		TourID:   request.TourID,
		TourName: request.TourName, // Ovo će se proslijediti iz request-a
		Price:    request.Price,    // Ovo će se proslijediti iz request-a
		Quantity: request.Quantity,
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

	return &models.ShoppingCartResponse{
		ID:         cart.ID,
		TouristID:  cart.TouristID,
		TotalPrice: cart.TotalPrice,
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
		total += item.Price * float64(item.Quantity)
	}

	return s.cartRepo.UpdateCartTotal(cartID, total)
}


