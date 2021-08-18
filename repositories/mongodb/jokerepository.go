package mongodb

import (
	"context"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type JokeRepository struct {
	c *mongo.Collection
}

func NewJokeRepository(c *mongo.Collection) *JokeRepository {
	return &JokeRepository{
		c: c,
	}
}

func (jr *JokeRepository) CheckValidID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	return primitive.IsValidObjectID(id)
}

func (jr *JokeRepository) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Jokes, *string, error) {
	var jokes data.Jokes

	objectID, err := primitive.ObjectIDFromHex(offset)

	if err != nil && offset != "" {
		return jokes, nil, repositories.ErrInvalidOffset
	}

	condition, options := paginate(objectID, "_id", limit+1, direction)

	cursor, err := jr.c.Find(ctx, condition, options)

	if err != nil {
		return jokes, nil, repositories.ErrInvalidOffset
	}

	defer cursor.Close(ctx)

	i := uint64(0)

	jokes = make(data.Jokes, 0, limit)

	var joke *data.Joke

	for i != limit && cursor.Next(ctx) {
		joke = new(data.Joke)

		if err = cursor.Decode(joke); err != nil {
			return jokes, nil, err
		}

		jokes = append(jokes, joke)
		i++
	}

	var nextID *string

	if cursor.Next(ctx) && joke != nil {
		nextID = &joke.ID
	}

	return jokes, nextID, nil
}

func (jr *JokeRepository) FetchOne(ctx context.Context, id string) (*data.Joke, error) {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, repositories.ErrInvalidOffset
	}

	result := jr.c.FindOne(ctx, bson.M{"_id": objectID})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, repositories.ErrUnknownID
		}

		return nil, result.Err()
	}

	joke := new(data.Joke)

	if err = result.Decode(joke); err != nil {
		return nil, err
	}

	return joke, nil
}

func (jr *JokeRepository) Insert(ctx context.Context, joke *data.Joke) (string, error) {
	result, err := jr.c.InsertOne(ctx, joke)

	if err != nil {
		return "", err
	}

	objectID := result.InsertedID.(primitive.ObjectID)

	return objectID.Hex(), nil
}

func (jr *JokeRepository) Update(ctx context.Context, id string, joke *data.Joke) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return "", err
	}

	result, err := jr.c.ReplaceOne(ctx, bson.M{"_id": objectID}, joke)

	if err != nil {
		return "", err
	}

	if result.MatchedCount == 0 {
		return objectID.Hex(), repositories.ErrUnknownID
	}

	return objectID.Hex(), nil
}

func (jr *JokeRepository) Delete(ctx context.Context, id string) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return "", err
	}

	result, err := jr.c.DeleteOne(ctx, bson.M{"_id": objectID})

	if err != nil {
		return "", err
	}

	if result.DeletedCount == 0 {
		return objectID.Hex(), repositories.ErrUnknownID
	}

	return objectID.Hex(), nil
}
