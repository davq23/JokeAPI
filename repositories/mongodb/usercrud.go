package mongodb

import (
	"context"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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

	user := new(data.User)

	if err = result.Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (jr *UserCRUD) Insert(ctx context.Context, user *data.User) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(*user.Password), 10)

	if err != nil {
		return "", err
	}

	password := string(passwordBytes)

	user.Password = &password

	result, err := jr.c.InsertOne(ctx, user)

	if err != nil {
		return "", err
	}

	objectID := result.InsertedID.(primitive.ObjectID)

	return objectID.Hex(), nil
}

func (jr *UserCRUD) Update(ctx context.Context, id string, user *data.User) (string, error) {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return "", err
	}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(*user.Password), 10)

	if err != nil {
		return "", err
	}

	password := string(passwordBytes)

	user.Password = &password

	result, err := jr.c.ReplaceOne(ctx, bson.M{"_id": objectID}, user)

	if err != nil {
		return "", err
	}

	if result.MatchedCount == 0 {
		return objectID.Hex(), repositories.ErrUnknownID
	}

	return objectID.Hex(), nil
}

func (jr *UserCRUD) Delete(ctx context.Context, id string) (string, error) {
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
