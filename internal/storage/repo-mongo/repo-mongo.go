package repomongo

import (
	"context"
	"errors"
	"fmt"
	"test-auth/internal/storage"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UsersDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewUsersDB(nameDB, collection, host string) (*UsersDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", host)))
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &UsersDB{
		client:     client,
		collection: client.Database(nameDB).Collection(collection),
	}, nil
}

func (db *UsersDB) Close(ctx context.Context) {
	if err := db.client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func (db *UsersDB) SetSession(ctx context.Context, userID string, session storage.Session) error {

	filter := bson.M{"guid": userID}

	newSession := bson.D{
		{Key: "refreshToken", Value: session.RefreshToken},
		{Key: "expiresAt", Value: session.ExpiresAt},
		{Key: "key", Value: session.Key},
	}

	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "sessions", Value: newSession},
		}},
	}
	_, err := db.collection.UpdateOne(ctx, filter, update)
	return err
}

func (db *UsersDB) GetUser(ctx context.Context, guid string) (storage.User, error) {
	var user storage.User
	if err := db.collection.FindOne(ctx, bson.M{"guid": guid}).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return storage.User{}, storage.ErrNotFound
		}

		return storage.User{}, err
	}

	return user, nil
}
