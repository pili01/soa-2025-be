package grpc_handlers

import (
	"context"
	"strings"
	"time"
	"tours-service/internal/models"
	"tours-service/internal/services"
	pb "tours-service/proto/compiled"

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

func (s *TourGRPCServer) CreateTour(ctx context.Context, req *pb.CreateTourRequest) (*pb.TourResponse, error) {
	if req.Tour == nil {
		return nil, status.Error(codes.InvalidArgument, "Tour data is required.")
	}
	if req.Keypoints == nil {
		return nil, status.Error(codes.InvalidArgument, "Keypoints data is required.")
	}

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
		}
		if strings.Contains(err.Error(), "at least two keypoints") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "Failed to create tour: "+err.Error())
	}

	res := &pb.TourResponse{
		Id:           int32(tour.ID),
		AuthorId:     int32(tour.AuthorID),
		Name:         tour.Name,
		Description:  tour.Description,
		Difficulty:   string(tour.Difficulty),
		Tags:         tour.Tags,
		Status:       string(tour.Status),
		Price:        tour.Price,
		DrivingStats: &pb.DistanceAndDuration{Distance: tour.DrivingStats.Distance, Duration: tour.DrivingStats.Duration},
		WalkingStats: &pb.DistanceAndDuration{Distance: tour.WalkingStats.Distance, Duration: tour.WalkingStats.Duration},
		CyclingStats: &pb.DistanceAndDuration{Distance: tour.CyclingStats.Distance, Duration: tour.CyclingStats.Duration},
	}

	if tour.TimePublished != nil {
		res.TimePublished = tour.TimePublished.Format(time.RFC3339)
	}
	if tour.TimeArchived != nil {
		res.TimeArchived = tour.TimeArchived.Format(time.RFC3339)
	}
	if tour.TimeDrafted != nil {
		res.TimeDrafted = tour.TimeDrafted.Format(time.RFC3339)
	}

	return res, nil
}

func (s *TourGRPCServer) GetToursByAuthorID(ctx context.Context, req *pb.GetToursByAuthorIDRequest) (*pb.GetToursByAuthorIDResponse, error) {
	tours, err := s.tourService.GetToursByAuthorID(int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to retrieve tours: "+err.Error())
	}

	var pbTours []*pb.TourResponse
	for _, tour := range tours {
		pbTour := &pb.TourResponse{
			Id:           int32(tour.ID),
			AuthorId:     int32(tour.AuthorID),
			Name:         tour.Name,
			Description:  tour.Description,
			Difficulty:   string(tour.Difficulty),
			Tags:         tour.Tags,
			Status:       string(tour.Status),
			Price:        tour.Price,
			DrivingStats: &pb.DistanceAndDuration{Distance: tour.DrivingStats.Distance, Duration: tour.DrivingStats.Duration},
			WalkingStats: &pb.DistanceAndDuration{Distance: tour.WalkingStats.Distance, Duration: tour.WalkingStats.Duration},
			CyclingStats: &pb.DistanceAndDuration{Distance: tour.CyclingStats.Distance, Duration: tour.CyclingStats.Duration},
		}

		if tour.TimePublished != nil {
			pbTour.TimePublished = tour.TimePublished.Format(time.RFC3339)
		}
		if tour.TimeArchived != nil {
			pbTour.TimeArchived = tour.TimeArchived.Format(time.RFC3339)
		}
		if tour.TimeDrafted != nil {
			pbTour.TimeDrafted = tour.TimeDrafted.Format(time.RFC3339)
		}

		pbTours = append(pbTours, pbTour)
	}

	return &pb.GetToursByAuthorIDResponse{
		Tours: pbTours,
	}, nil
}
