package mongodb

import (
	"github.com/davq23/jokeapi/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func paginate(offset interface{}, cursorName string, limit uint64, direction repositories.FetchDirection) (bson.M, *options.FindOptions) {
	options := options.Find()

	var filterCondition string

	switch direction {
	case repositories.FetchNext:
		filterCondition = "$gte"
	case repositories.FetchBack:
		filterCondition = "$lt"
	}

	options.SetSort(bson.M{cursorName: direction})
	options.SetLimit(int64(limit))

	return bson.M{cursorName: bson.M{filterCondition: offset}}, options
}
