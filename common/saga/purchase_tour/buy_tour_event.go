package purchase_tour

type BuyTourItem struct {
	TourID   int `json:"tour_id"`
	Quantity int `json:"quantity"`
}

type BuyTourDetails struct {
	PurchaseID string        `json:"purchase_id"`
	CartID     int           `json:"cart_id"`
	TouristID  int           `json:"tourist_id"`
	Items      []BuyTourItem `json:"items"`
}

type BuyTourCommandType int8

const (
	ReserveCapacity BuyTourCommandType = iota
	ReleaseCapacity
	IssueTokens
	CompletePurchase
	AbortPurchase
	UnknownCommand
)

type BuyTourCommand struct {
	Purchase BuyTourDetails     `json:"purchase"`
	Type     BuyTourCommandType `json:"type"`
}

type BuyTourReplyType int8

const (
	CapacityReserved BuyTourReplyType = iota
	CapacityNotReserved
	CapacityReleased
	CapacityReleaseFailed

	TokensIssued
	TokensNotIssued

	PurchaseCompleted
	PurchaseAborted

	UnknownReply
)

type BuyTourReply struct {
	Purchase BuyTourDetails   `json:"purchase"`
	Type     BuyTourReplyType `json:"type"`

	FailedTourIDs []int  `json:"failed_tour_ids,omitempty"`
	Message       string `json:"message,omitempty"`
}
