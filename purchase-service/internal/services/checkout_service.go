package services

import (
	"errors"
	"log"

	"purchase-service/internal/models"
	"purchase-service/internal/repositories"
	"github.com/google/uuid"
)

type CheckoutService struct {
	cartRepo      *repositories.ShoppingCartRepository
	itemRepo      *repositories.OrderItemRepository
	tokenRepo     *repositories.TourPurchaseTokenRepository
}

func NewCheckoutService(cartRepo *repositories.ShoppingCartRepository, itemRepo *repositories.OrderItemRepository, tokenRepo *repositories.TourPurchaseTokenRepository) *CheckoutService {
	return &CheckoutService{
		cartRepo:    cartRepo,
		itemRepo:    itemRepo,
		tokenRepo:   tokenRepo,
	}
}

func (s *CheckoutService) ProcessCheckout(touristID int, request *models.CheckoutRequest) (*models.CheckoutResponse, error) {
	
	cart, err := s.cartRepo.GetCartByTouristID(touristID)
	if err != nil {
		return nil, err
	}

	if cart == nil {
		return nil, errors.New("cart not found")
	}

	if cart.ID != request.CartID {
		return nil, errors.New("unauthorized cart access")
	}

	
	items, err := s.itemRepo.GetItemsByCartID(cart.ID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errors.New("cart is empty")
	}

	
	var tokens []models.TourPurchaseToken
	for _, item := range items {
		tokenValue := uuid.New().String()
		token := &models.TourPurchaseToken{
			TouristID: touristID,
			TourID:    item.TourID,
			Token:     tokenValue,
		}

	
		err = s.tokenRepo.CreateToken(token)
		if err != nil {
			log.Printf("Error creating token for tour %d: %v", item.TourID, err)
			continue
		}

		tokens = append(tokens, *token)
	}

	if len(tokens) == 0 {
		return nil, errors.New("no tours could be purchased")
	}

	
	err = s.itemRepo.ClearCart(cart.ID)
	if err != nil {
		log.Printf("Error clearing cart: %v", err)
	}

	
	err = s.cartRepo.UpdateCartTotal(cart.ID, 0.0)
	if err != nil {
		log.Printf("Error resetting cart total: %v", err)
	}

	return &models.CheckoutResponse{
		Success: true,
		Tokens:  tokens,
		Message: "Checkout completed successfully",
	}, nil
}

func (s *CheckoutService) GetPurchaseHistory(touristID int) (*models.PurchaseHistoryResponse, error) {
	tokens, err := s.tokenRepo.GetTokensByTouristID(touristID)
	if err != nil {
		return nil, err
	}

	return &models.PurchaseHistoryResponse{
		TouristID: touristID,
		Purchases: tokens,
	}, nil
}

func (s *CheckoutService) ValidateToken(tokenValue string, tourID int) (bool, error) {
	return s.tokenRepo.ValidateToken(tokenValue, tourID)
}
