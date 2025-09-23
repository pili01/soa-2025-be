package services

import (
	"log"
	"purchase-service/internal/models"

	saga "example.com/common/saga/messaging"
	events "example.com/common/saga/purchase_tour"
)

type TourPurchaseOrchestrator struct {
	commandPublisher saga.Publisher
	replySubscriber  saga.Subscriber
}

func NewTourPurchaseOrchestrator(publisher saga.Publisher, subscriber saga.Subscriber) (*TourPurchaseOrchestrator, error) {
	o := &TourPurchaseOrchestrator{
		commandPublisher: publisher,
		replySubscriber:  subscriber,
	}
	if err := o.replySubscriber.Subscribe(o.handle); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *TourPurchaseOrchestrator) Start(purchaseID string, cart models.ShoppingCartResponse) error {
	cmd := &events.BuyTourCommand{
		Type: events.ReserveCapacity,
		Purchase: events.BuyTourDetails{
			PurchaseID: purchaseID,
			CartID:     cart.ID,
			TouristID:  cart.TouristID,
			Items:      make([]events.BuyTourItem, 0, len(cart.Items)),
		},
	}
	for _, it := range cart.Items {
		cmd.Purchase.Items = append(cmd.Purchase.Items, events.BuyTourItem{
			TourID:   it.TourID,
			Quantity: it.Quantity,
		})
	}
	return o.commandPublisher.Publish(cmd)
}

func (o *TourPurchaseOrchestrator) handle(reply *events.BuyTourReply) {

	next := o.nextCommandType(reply.Type)
	if next == events.UnknownCommand {
		log.Printf("[Orchestrator] Saga for purchase %s has finished or been aborted.", reply.Purchase.PurchaseID) // <-- DODAJTE
		return
	}
	if next == events.UnknownCommand {
		return
	}

	cmd := &events.BuyTourCommand{
		Purchase: reply.Purchase,
		Type:     next,
	}
	log.Printf("[Orchestrator] Sending next command '%s' for purchase %s", cmd.Type, cmd.Purchase.PurchaseID)
	_ = o.commandPublisher.Publish(cmd)
}

func (o *TourPurchaseOrchestrator) nextCommandType(rt events.BuyTourReplyType) events.BuyTourCommandType {
	switch rt {
	// HAPPY PATH
	case events.CapacityReserved:
		return events.IssueTokens
	case events.TokensIssued:
		return events.CompletePurchase

	// FAIL GRANE
	case events.CapacityNotReserved:
		return events.AbortPurchase
	case events.TokensNotIssued:
		return events.ReleaseCapacity
	case events.CapacityReleased:
		return events.AbortPurchase
	case events.CapacityReleaseFailed:
		return events.AbortPurchase

	// TERMINAL
	case events.PurchaseCompleted, events.PurchaseAborted:
		return events.UnknownCommand

	default:
		return events.UnknownCommand
	}
}
