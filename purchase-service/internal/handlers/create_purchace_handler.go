package handlers

import (
	"context"
	"log"

	"purchase-service/internal/services"

	saga "example.com/common/saga/messaging"
	events "example.com/common/saga/purchase_tour"
)

type CreatePurchaseCommandHandler struct {
	checkoutService   *services.CheckoutService
	replyPublisher    saga.Publisher
	commandSubscriber saga.Subscriber
}

func NewCreatePurchaseCommandHandler(
	checkoutService *services.CheckoutService,
	replyPublisher saga.Publisher,
	commandSubscriber saga.Subscriber,
) (*CreatePurchaseCommandHandler, error) {

	h := &CreatePurchaseCommandHandler{
		checkoutService:   checkoutService,
		replyPublisher:    replyPublisher,
		commandSubscriber: commandSubscriber,
	}

	if err := h.commandSubscriber.Subscribe(h.handle); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *CreatePurchaseCommandHandler) handle(command *events.BuyTourCommand) {
	reply := events.BuyTourReply{Purchase: command.Purchase}

	switch command.Type {

	case events.IssueTokens:
		// izdaj tokene za sve stavke u kupovini
		err := h.checkoutService.IssueTokensForItems(
			context.Background(),
			command.Purchase.PurchaseID,
			command.Purchase.TouristID,
			command.Purchase.Items,
		)
		if err != nil {
			log.Printf("[Purchase-Handler] ERROR issuing tokens for purchase %s: %v", command.Purchase.PurchaseID, err)
			reply.Type = events.TokensNotIssued
			reply.Message = err.Error()
			_ = h.replyPublisher.Publish(&reply)
			return
		}

		for _, item := range command.Purchase.Items {
			//err := h.toursServiceHttpClient.MarkTourAsPurchased(item.TourID, command.Purchase.TouristID)
			if err != nil {
				log.Printf("ERROR: Failed to notify tours-service about purchase of tour %d by tourist %d", item.TourID, command.Purchase.TouristID)
			}
		}
		reply.Type = events.TokensIssued

	case events.CompletePurchase:
		h.checkoutService.ClearCart(command.Purchase.CartID)
		reply.Type = events.PurchaseCompleted

	default:
		reply.Type = events.UnknownReply
	}

	if reply.Type != events.UnknownReply {
		_ = h.replyPublisher.Publish(&reply)
	}
}
