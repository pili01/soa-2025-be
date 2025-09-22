package services

import (
	"purchase-service/internal/models"

	events "github.com/tamararankovic/microservices_demo/common/saga/create_order"
	saga "github.com/tamararankovic/microservices_demo/common/saga/messaging"
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
		Purchase: events.BuyTourDetails{
			PurchaseID: purchaseID,
			CartID:     cart.ID,
			TouristID:  cart.TouristID,
			Items:      make([]events.BuyTourItem, 0, len(cart.Items)),
		},
		Type: events.ReserveCapacity,
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
		return
	}

	cmd := &events.BuyTourCommand{
		Purchase: reply.Purchase,
		Type:     next,
	}
	_ = o.commandPublisher.Publish(cmd)
}

func (o *TourPurchaseOrchestrator) nextCommandType(reply events.BuyTourReplyType) events.BuyTourCommandType {
	switch reply {

	case events.CapacityReserved:
		return events.IssueTokens

	case events.CapacityNotReserved:
		return events.AbortPurchase

	case events.CapacityReleased:
		return events.AbortPurchase

	case events.CapacityReleaseFailed:
		return events.AbortPurchase

	case events.TokensIssued:
		return events.CompletePurchase

	case events.TokensNotIssued:
		return events.ReleaseCapacity

	case events.PurchaseCompleted, events.PurchaseAborted:
		return events.UnknownCommand

	default:
		return events.UnknownCommand
	}
}
