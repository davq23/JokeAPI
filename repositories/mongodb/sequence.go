package mongodb

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sequence struct {
	ID       primitive.ObjectID `bson:"_id"`
	Sequence uint64             `bson:"seq"`
}
