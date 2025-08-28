package services

import(
	"tours-service/internal/repositories"
	"tours-service/internal/models"
)

type TourReviewService struct{
	TourReviewRepository *repositories.TourReviewRepository
}

func NewTourReviewService(tourReviewRepository *repositories.TourReviewRepository) *TourReviewService{
	return &TourReviewService{TourReviewRepository: tourReviewRepository}
}

func (t *TourReviewService) CreateTourReview(tourReview *models.TourReview) error{
	return t.TourReviewRepository.CreateTourReview(tourReview);
}

func (t *TourReviewService) GetReviewsByTourID(tourReviewId int) ([]models.TourReview, error) {
	return t.TourReviewRepository.GetReviewsByTourID(tourReviewId);
}