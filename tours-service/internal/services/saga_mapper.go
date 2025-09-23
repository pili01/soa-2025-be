package services

import (
	"tours-service/internal/models"

	events "example.com/common/saga/purchase_tour"
)

func mapReserveCapacity(command *events.BuyTourCommand) map[*models.TourCapacity]int {
	capacities := make(map[*models.TourCapacity]int)
	for _, item := range command.Purchase.Items {
		capacity := &models.TourCapacity{
			TourID: item.TourID,
		}
		capacities[capacity] = -item.Quantity
	}
	return capacities
}

func mapReleaseCapacity(command *events.BuyTourCommand) map[*models.TourCapacity]int {
	capacities := make(map[*models.TourCapacity]int)
	for _, item := range command.Purchase.Items {
		capacity := &models.TourCapacity{
			TourID: item.TourID,
		}
		capacities[capacity] = item.Quantity
	}
	return capacities
}
