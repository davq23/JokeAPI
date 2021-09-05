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

type UserCRUD struct {
	c *mongo.Collection
}

func NewUser(c *mongo.Collection) *UserCRUD {
	return &UserCRUD{
		c: c,
	}
}

func (jr *UserCRUD) CheckValidID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	return primitive.IsValidObjectID(id)
}

func (jr *UserCRUD) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Users, *string, error) {
	var jokes data.Users

	condition, options := paginate(offset, "id", limit+1, direction)

	cursor, err := jr.c.Find(ctx, condition, options)

	if err != nil {
		return jokes, nil, repositories.ErrInvalidOffset
	}

	defer cursor.Close(ctx)

	i := uint64(0)

	jokes = make(data.Users, 0, limit)

	var joke *data.User

	for i != limit && cursor.Next(ctx) {
		joke = new(data.User)

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

func (jr *UserCRUD) FetchOne(ctx context.Context, id string) (*data.User, error) {
	result := jr.c.FindOne(ctx, bson.M{"id": id})

	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repositories.ErrUnknownID
		}

		return nil, err
	}

	user := new(data.User)

	if err := result.Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (jr *UserCRUD) FetchOneByEmail(ctx context.Context, email string) (*data.User, error) {
	result := jr.c.FindOne(ctx, bson.M{"email": email})

	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, repositories.ErrUnknownEmail
		}

		return nil, err
	}

	user := new(data.User)

	if err := result.Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (jr *UserCRUD) Insert(ctx context.Context, user *data.User) (string, error) {
	result, err := jr.c.InsertOne(ctx, user)

	if err != nil {
		return "", err
	}

	objectID := result.InsertedID.(primitive.ObjectID)

	return objectID.Hex(), nil
}

func (jr *UserCRUD) Update(ctx context.Context, id string, user *data.User) (string, error) {
	result, err := jr.c.ReplaceOne(ctx, bson.M{"id": id}, user)

	if err != nil {
		return "", err
	}

	if result.MatchedCount == 0 {
		return id, repositories.ErrUnknownID
	}

	return id, nil
}

func (jr *UserCRUD) Delete(ctx context.Context, id string) (string, error) {
	result, err := jr.c.DeleteOne(ctx, bson.M{"id": id})

	if err != nil {
		return "", err
	}

	if result.DeletedCount == 0 {
		return id, repositories.ErrUnknownID
	}

	return id, nil
}
