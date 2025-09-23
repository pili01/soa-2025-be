package services

import (
	"context"
	"time"

	events "example.com/common/saga/purchase_tour"

	"purchase-service/internal/models"
	"purchase-service/internal/repositories"

	"github.com/google/uuid"
)

type CheckoutService struct {
	cartRepo    *repositories.ShoppingCartRepository
	itemRepo    *repositories.OrderItemRepository
	tokenRepo   *repositories.TourPurchaseTokenRepository
	cartService *CartService
}

func (s *CheckoutService) IssueTokensForItems(
	ctx context.Context,
	purchaseID string,
	touristID int,
	items []events.BuyTourItem,
) error {
	for _, it := range items {
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		for i := 0; i < qty; i++ {
			token := &models.TourPurchaseToken{
				TouristID:   touristID,
				TourID:      it.TourID,
				Token:       uuid.New().String(),
				PurchasedAt: time.Now(),
			}
			if err := s.tokenRepo.CreateToken(token); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewCheckoutService(cartRepo *repositories.ShoppingCartRepository, itemRepo *repositories.OrderItemRepository, tokenRepo *repositories.TourPurchaseTokenRepository, cartService *CartService) *CheckoutService {
	return &CheckoutService{
		cartRepo:    cartRepo,
		itemRepo:    itemRepo,
		tokenRepo:   tokenRepo,
		cartService: cartService,
	}
}

/*func (s *CheckoutService) ProcessCheckout(touristID int, request *models.CheckoutRequest) (*models.CheckoutResponse, error) {

	cart, err := s.cartRepo.GetCartByTouristID(touristID)
	if err != nil {
		return nil, err
	}

	if cart == nil {
		return nil, errors.New("cart not found")
	}

	fmt.Println("Cart ID:", cart.ID)
	fmt.Println("Request Cart ID:", request.CartID)
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

	// Ne resetujemo total_price - ostaje kao je bio
	// err = s.cartRepo.UpdateCartTotal(cart.ID, 0.0)

	return &models.CheckoutResponse{
		Success: true,
		Tokens:  tokens,
		Message: "Checkout completed successfully",
	}, nil
}*/

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

func (s *CheckoutService) CheckIsPurchased(touristID int, tourID int) (bool, error) {
	return s.tokenRepo.CheckIsPurchased(touristID, tourID)
}

func (s *CheckoutService) ClearCart(cartID int) error {
	// itemRepo je već dostupan u CheckoutService strukturi
	return s.itemRepo.ClearCart(cartID)
}

func (s *CheckoutService) GetCartForUser(touristID int) (*models.ShoppingCartResponse, error) {
	// Poziva metodu iz servisa koji je za to zadužen.
	return s.cartService.GetCart(touristID)
}
