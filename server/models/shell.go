package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shell struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID string             `json:"user_id" bson:"user_id"`
	UUID   string             `json:"uuid" bson:"uuid"`
}
