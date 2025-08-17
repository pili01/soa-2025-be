package grpc_handlers

import (
	"context"
	"strings"
	"tours-service/internal/models"
	"tours-service/internal/services"
	pb "tours-service/proto-files/tours" 
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TourGRPCServer struct {
	pb.UnimplementedTourServiceServer
	tourService *services.TourService
}

func NewTourGRPCServer(tourService *services.TourService) *TourGRPCServer {
	return &TourGRPCServer{
		tourService: tourService,
	}
}

func (s *TourGRPCServer) CreateTour(ctx context.Context, req *pb.CreateTourRequest) (*pb.CreateTourResponse, error) {
	tour := &models.Tour{
		Name:        req.Tour.Name,
		Description: req.Tour.Description,
		Difficulty:  models.TourDifficulty(req.Tour.Difficulty),
		Tags:        req.Tour.Tags,
		AuthorID:    int(req.UserId),
		Status:      models.StatusDraft, 
		Price:       0.0, 
	}

	var keypoints []*models.Keypoint
	for _, kp := range req.Keypoints {
		keypoints = append(keypoints, &models.Keypoint{
			Name:        kp.Name,
			Description: kp.Description,
			ImageURL:    kp.ImageUrl,
			Latitude:    kp.Latitude,
			Longitude:   kp.Longitude,
			Ordinal:     int(kp.Ordinal),
		})
	}
	err := s.tourService.CreateTour(tour, keypoints)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else if strings.Contains(err.Error(), "at least two keypoints") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			return nil, status.Error(codes.Internal, "Failed to create tour: "+err.Error())
		}
	}

	return &pb.CreateTourResponse{
		Id: int32(tour.ID),
	}, nil
}