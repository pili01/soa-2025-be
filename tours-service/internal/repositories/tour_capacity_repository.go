package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tours-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrCapacityNotFound = errors.New("tour capacity not found")
	ErrNotEnoughSeats   = errors.New("not enough available seats")
)

type TourCapacityRepository struct {
	Collection *mongo.Collection
}

func NewTourCapacityRepository(db *mongo.Database) *TourCapacityRepository {
	return &TourCapacityRepository{
		Collection: db.Collection("tour_capacities"),
	}
}

func (r *TourCapacityRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.Collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tourId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *TourCapacityRepository) InitCapacity(tourID, capacity int) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    now := time.Now()

    var existing models.TourCapacity
    err := r.Collection.FindOne(ctx, bson.M{"tourId": tourID}).Decode(&existing)
    if err == mongo.ErrNoDocuments {
        doc := models.TourCapacity{
            TourID:         tourID,
            Capacity:       capacity,
            AvailableSeats: capacity,
            UpdatedAt:      now,
        }
        if _, err := r.Collection.InsertOne(ctx, doc); err != nil {
            return fmt.Errorf("create capacity: %w", err)
        }
        return nil
    }
    if err != nil {
        return fmt.Errorf("read capacity: %w", err)
    }

    update := bson.M{
        "$set": bson.M{
            "capacity":  capacity,
            "updatedAt": now,
        },
    }
    if existing.AvailableSeats > capacity {
        update["$set"].(bson.M)["availableSeats"] = capacity
    }

    if _, err := r.Collection.UpdateOne(ctx, bson.M{"tourId": tourID}, update); err != nil {
        return fmt.Errorf("update capacity: %w", err)
    }
    return nil
}


func (r *TourCapacityRepository) GetCapacity(tourID int) (*models.TourCapacity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cap models.TourCapacity
	if err := r.Collection.FindOne(ctx, bson.M{"tourId": tourID}).Decode(&cap); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCapacityNotFound
		}
		return nil, fmt.Errorf("get capacity: %w", err)
	}
	return &cap, nil
}

func (r *TourCapacityRepository) ConsumeSeats(tourID, qty int) error {
	if qty <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	filter := bson.M{
		"tourId":         tourID,
		"availableSeats": bson.M{"$gte": qty},
	}
	update := bson.M{
		"$inc": bson.M{"availableSeats": -qty},
		"$set": bson.M{"updatedAt": now},
	}

	res := r.Collection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			_, getErr := r.GetCapacity(tourID)
			if getErr == ErrCapacityNotFound {
				return ErrCapacityNotFound
			}
			return ErrNotEnoughSeats
		}
		return fmt.Errorf("consume seats: %w", err)
	}
	return nil
}

func (r *TourCapacityRepository) ReleaseSeats(tourID, qty int) error {
	if qty <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	update := mongo.Pipeline{
		{{
			Key: "$set", Value: bson.M{
				"availableSeats": bson.M{
					"$min": bson.A{
						bson.M{"$add": bson.A{"$availableSeats", qty}},
						"$capacity",
					},
				},
				"updatedAt": now,
			},
		}},
	}

	res := r.Collection.FindOneAndUpdate(ctx, bson.M{"tourId": tourID}, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCapacityNotFound
		}
		return fmt.Errorf("release seats: %w", err)
	}
	return nil
}
