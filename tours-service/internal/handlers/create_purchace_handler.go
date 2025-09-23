package handlers

import (
	"fmt"
	"log"

	"tours-service/internal/services"

	saga "example.com/common/saga/messaging"
	events "example.com/common/saga/purchase_tour"
)

type CreatePurchaseCommandHandler struct {
	capacityService   *services.TourCapacityService
	replyPublisher    saga.Publisher
	commandSubscriber saga.Subscriber
}

func NewCreatePurchaseCommandHandler(
	capacityService *services.TourCapacityService,
	replyPublisher saga.Publisher,
	commandSubscriber saga.Subscriber,
) (*CreatePurchaseCommandHandler, error) {

	h := &CreatePurchaseCommandHandler{
		capacityService:   capacityService,
		replyPublisher:    replyPublisher,
		commandSubscriber: commandSubscriber,
	}
	if err := h.commandSubscriber.Subscribe(h.handle); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *CreatePurchaseCommandHandler) handle(cmd *events.BuyTourCommand) {
	reply := events.BuyTourReply{Purchase: cmd.Purchase}

	switch cmd.Type {

	case events.ReserveCapacity:
		items := foldItems(cmd)
		if err := h.reserveAll(items); err != nil {
			log.Printf("[Tours-Handler] ERROR reserving capacity for purchase %s: %v", cmd.Purchase.PurchaseID, err)
			reply.Type = events.CapacityNotReserved
			reply.Message = err.Error()
			_ = h.replyPublisher.Publish(&reply)
			return
		}
		reply.Type = events.CapacityReserved

	case events.ReleaseCapacity:
		items := foldItems(cmd)
		if err := h.releaseAll(items); err != nil {
			reply.Type = events.CapacityReleaseFailed
			reply.Message = err.Error()
			_ = h.replyPublisher.Publish(&reply)
			return
		}
		reply.Type = events.CapacityReleased

	default:
		reply.Type = events.UnknownReply
		log.Printf("[Tours-Handler] Sending reply '%s' for purchase %s", reply.Type, cmd.Purchase.PurchaseID)
	}

	if reply.Type != events.UnknownReply {
		_ = h.replyPublisher.Publish(&reply)
	}
}

func foldItems(cmd *events.BuyTourCommand) map[int]int {
	out := make(map[int]int, len(cmd.Purchase.Items))
	for _, it := range cmd.Purchase.Items {
		out[it.TourID] += it.Quantity
	}
	return out
}

func (h *CreatePurchaseCommandHandler) reserveAll(m map[int]int) error {
	consumed := make(map[int]int, len(m))

	for tourID, qty := range m {
		if qty <= 0 {
			continue
		}
		if err := h.capacityService.Consume(tourID, qty); err != nil {
			for tID, cqty := range consumed {
				_ = h.capacityService.Release(tID, cqty)
			}
			return fmt.Errorf("reserve failed for tour %d: %w", tourID, err)
		}
		consumed[tourID] = qty
	}
	return nil
}

func (h *CreatePurchaseCommandHandler) releaseAll(m map[int]int) error {
	for tourID, qty := range m {
		if qty <= 0 {
			continue
		}
		if err := h.capacityService.Release(tourID, qty); err != nil {
			return fmt.Errorf("release failed for tour %d: %w", tourID, err)
		}
	}
	return nil
}
