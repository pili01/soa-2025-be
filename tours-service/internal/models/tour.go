package models

type TourStatus string
type TourDifficulty string

const (
    StatusDraft     TourStatus = "Draft"
    StatusPublished TourStatus = "Published"
    StatusArchived  TourStatus = "Archived"
)

const (
    DifficultyEasy   TourDifficulty = "Easy"
    DifficultyMedium TourDifficulty = "Medium"
    DifficultyHard   TourDifficulty = "Hard"
)

type Tour struct {
	ID int `bson:"_id,omitempty" json:"id"`
	AuthorID int `bson:"authorId" json:"authorId"`
	Name string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Difficulty TourDifficulty `bson:"difficulty" json:"difficulty"` // Easy, Medium, Hard
	Tags []string `bson:"tags" json:"tags"`
	Status TourStatus `bson:"status" json:"status"` // Draft, Published, Archived
	Price float64 `bson:"price" json:"price"`
	Keypoints []Keypoint `bson:"keypoints,omitempty" json:"keypoints,omitempty"` // Lista keypoint-a
	Reviews []TourReview `bson:"reviews,omitempty" json:"reviews,omitempty"`
}