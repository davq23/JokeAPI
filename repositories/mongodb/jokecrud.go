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

type JokeCRUD struct {
	c *mongo.Collection
}

func NewJoke(c *mongo.Collection) *JokeCRUD {
	return &JokeCRUD{
		c: c,
	}
}

func (jr *JokeCRUD) CheckValidID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	return primitive.IsValidObjectID(id)
}

func (jr *JokeCRUD) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Jokes, *string, error) {
	var jokes data.Jokes

	condition, _ := paginate(offset, "id", limit+1, direction)

	cursor, err := jr.c.Aggregate(ctx,
		bson.A{bson.M{"$match": condition},

			bson.M{"$unwind": bson.M{"path": "$ratings", "preserveNullAndEmptyArrays": true}},
			bson.M{
				"$group": bson.M{
					"_id":         "$id",
					"id":          bson.M{"$first": "$id"},
					"text":        bson.M{"$first": "$text"},
					"explanation": bson.M{"$first": "$explanation"},
					"language":    bson.M{"$first": "$language"},
					"avgRating": bson.M{
						"$avg": bson.M{"$cond": bson.A{
							bson.M{
								"$eq": bson.A{"$ratings", nil},
							}, 0, "$ratings.rating",
						}},
					},
				},
			},
			bson.M{
				"$sort": bson.M{
					"id": direction,
				},
			},
			bson.M{
				"$limit": limit + 1,
			},
		})

	if err != nil {
		return jokes, nil, err
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
		joke := new(data.Joke)

		if err = cursor.Decode(joke); err != nil {
			return jokes, nil, err
		}

		nextID = &joke.ID
	}

	return jokes, nextID, nil
}

func (jr *JokeCRUD) FetchRatings(ctx context.Context, jokeID string, limit uint64, offset string, direction repositories.FetchDirection) (data.JokeRatings, *string, error) {
	//, _ := paginate(offset, "ratings.id", limit+1, direction)
	return nil, nil, nil
}

func (jr *JokeCRUD) FetchOne(ctx context.Context, id string) (*data.Joke, error) {
	result := jr.c.FindOne(ctx, bson.M{"id": id})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, repositories.ErrUnknownID
		}

		return nil, result.Err()
	}

	joke := new(data.Joke)

	if err := result.Decode(joke); err != nil {
		return nil, err
	}

	return joke, nil
}

func (jr *JokeCRUD) Insert(ctx context.Context, joke *data.Joke) (string, error) {
	result, err := jr.c.InsertOne(ctx, joke)

	if err != nil {
		return "", err
	}

	objectID := result.InsertedID.(primitive.ObjectID)

	return objectID.Hex(), nil
}

func (jr *JokeCRUD) DeleteRating(ctx context.Context, jokeID string, ratingID string, authID string) (string, error) {
	return "", nil
}

func (jr *JokeCRUD) RateJoke(ctx context.Context, jokeID string, jokeRating *data.JokeRating) (string, error) {
	session, err := jr.c.Database().Client().StartSession()

	if err != nil {
		return "", err
	}

	defer session.EndSession(ctx)

	if err = session.StartTransaction(); err != nil {
		return "", err
	}

	result := jr.c.FindOne(ctx, bson.M{"id": jokeID})

	err = result.Err()

	if err != nil {
		session.AbortTransaction(ctx)

		if err == mongo.ErrNoDocuments {
			return "", repositories.ErrUnknownID
		}

		return "", err
	}

	_, err = jr.c.UpdateOne(ctx, bson.M{"id": jokeID}, bson.M{"$push": bson.M{
		"ratings": jokeRating,
	}})

	if err != nil {
		session.AbortTransaction(ctx)
		return "", err
	}

	session.CommitTransaction(ctx)

	return jokeRating.ID, nil
}

func (jr *JokeCRUD) Update(ctx context.Context, id string, joke *data.Joke) (string, error) {
	result, err := jr.c.ReplaceOne(ctx, bson.M{"id": id}, joke)

	if err != nil {
		return "", err
	}

	if result.MatchedCount == 0 {
		return id, repositories.ErrUnknownID
	}

	return id, nil
}

func (jr *JokeCRUD) Delete(ctx context.Context, id string) (string, error) {
	result, err := jr.c.DeleteOne(ctx, bson.M{"id": id})

	if err != nil {
		return "", err
	}

	if result.DeletedCount == 0 {
		return id, repositories.ErrUnknownID
	}

	return id, nil
}
