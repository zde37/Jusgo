package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Jusgo struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Joke      string             `bson:"joke" json:"joke"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
