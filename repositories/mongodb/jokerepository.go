package mongodb

import (
	"context"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
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

func (jr *JokeRepository) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Jokes, *string, error) {
	var jokes data.Jokes

	objectID, err := primitive.ObjectIDFromHex(offset)

	if err != nil && offset != "" {
		return jokes, nil, repositories.ErrInvalidOffset
	}

	condition, options := paginate(objectID, "_id", limit, direction)

	cursor, err := jr.c.Find(ctx, condition, options)

	if err != nil {
		return jokes, nil, repositories.ErrInvalidOffset
	}

	defer cursor.Close(ctx)

	i := uint64(limit)

	jokes = make(data.Jokes, 0, limit)

	var joke *data.Joke

	for cursor.Next(ctx) && i == limit {
		joke = new(data.Joke)

		if err = cursor.Decode(joke); err != nil {
			return jokes, nil, err
		}

		jokes = append(jokes, joke)
		i++
	}

	var nextID *string

	if joke != nil {
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

	if err != nil {
		return nil, err
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	joke := new(data.Joke)

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

	_, err = jr.c.UpdateByID(ctx, objectID, joke)

	if err != nil {
		return "", err
	}

	return objectID.Hex(), nil
}
