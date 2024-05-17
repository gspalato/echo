package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"unreal.sh/echo/internal/structures"
)

const UserCollectionName = "users"
const DisposalCollectionName = "disposals"

type DatabaseService struct {
	Client *mongo.Client

	dbUrl  string
	dbName string
	dbUser string
	dbPass string
}

func (ds *DatabaseService) Init(ctx context.Context) {
	dbUrl, found := os.LookupEnv("DATABASE_URI")
	if !found {
		panic("No database URI found in environment.")
	}
	ds.dbUrl = dbUrl

	dbName, found := os.LookupEnv("DATABASE_NAME")
	if !found {
		panic("No database name found in environment.")
	}
	ds.dbName = dbName

	dbUser, found := os.LookupEnv("DATABASE_USER")
	if !found {
		panic("No database user found in environment.")
	}
	ds.dbUser = dbUser

	dbPassword, found := os.LookupEnv("DATABASE_PASSWORD")
	if !found {
		panic("No database password found in environment.")
	}
	ds.dbPass = dbPassword

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(dbUrl).SetAuth(
			options.Credential{Username: dbUser, Password: dbPassword}))

	if err != nil {
		panic(err)
	}

	ds.Client = client

	fmt.Println("Database connected.")
}

func (ds *DatabaseService) GetUserById(id string) (*structures.User, error) {
	var result structures.User

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("Invalid ID.")
		return nil, structures.ErrInvalidDatabaseId
	}

	filter := bson.M{"_id": objectId}

	err = ds.Client.Database(ds.dbName).Collection(UserCollectionName).FindOne(
		context.Background(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, structures.ErrNoUser
	} else if err != nil {
		fmt.Printf("Failed to get user %v: %v\n", id, err)
		return nil, err
	}

	return &result, nil
}

func (ds *DatabaseService) GetUserByUsername(username string) (*structures.User, error) {
	var result structures.User

	filter := bson.M{"username": username}

	err := ds.Client.Database(ds.dbName).Collection(UserCollectionName).FindOne(
		context.Background(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, structures.ErrNoUser
	} else if err != nil {
		fmt.Printf("Failed to get user %v: %v\n", username, err)
		return nil, err
	}

	return &result, nil
}

func (ds *DatabaseService) CreateUser(user structures.User) error {
	_, err := ds.Client.Database(ds.dbName).Collection(UserCollectionName).InsertOne(context.Background(), user)
	if err != nil {
		fmt.Printf("Failed to create user %v: %v\n", user.Username, err)
		return err
	}

	fmt.Printf("Created user %v.\n", user.Username)

	return nil
}

func (ds *DatabaseService) UpdateUserById(id string, update interface{}) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("Invalid ID.")
		return structures.ErrInvalidDatabaseId
	}

	r, err := ds.Client.Database(ds.dbName).Collection(UserCollectionName).UpdateOne(context.Background(),
		primitive.M{"_id": objectId}, update)

	if err != nil {
		fmt.Printf("Failed to update user %v: %v\n", id, err)
		return err
	}

	if r.MatchedCount == 0 {
		fmt.Printf("No users matched the filter.\n")
		return structures.ErrNoUser
	}

	fmt.Printf("Updated user %v.\n", id)

	return nil
}

func (ds *DatabaseService) GetDisposalsByUserId(userId string) ([]structures.DisposalClaim, error) {
	var result *[]structures.DisposalClaim

	filter := bson.M{"user_id": userId}

	cur, err := ds.Client.Database(ds.dbName).Collection(DisposalCollectionName).Find(
		context.Background(), filter)

	if err != nil {
		fmt.Printf("Failed to get disposals for user %v: %v\n", userId, err)
		return nil, err
	}

	err = cur.All(context.Background(), &result)
	if err != nil {
		fmt.Printf("Failed to get disposals for user %v: %v\n", userId, err)
		return nil, err
	}

	return *result, nil
}

func (ds *DatabaseService) InsertDisposal(disposal *structures.DisposalClaim) error {
	_, err := ds.Client.Database(ds.dbName).Collection(DisposalCollectionName).InsertOne(context.Background(), disposal)
	if err != nil {
		fmt.Printf("Failed to insert disposal: %v\n", err)
		return err
	}

	fmt.Printf("Inserted disposal %v.\n", disposal.Token)

	return nil
}

func (ds *DatabaseService) GetDisposalByToken(token string) (*structures.DisposalClaim, error) {
	var result structures.DisposalClaim

	filter := bson.M{"token": token}

	err := ds.Client.Database(ds.dbName).Collection(DisposalCollectionName).FindOne(
		context.Background(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, structures.ErrNoDisposal
	} else if err != nil {
		fmt.Printf("Failed to get disposal %v: %v\n", token, err)
		return nil, err
	}

	return &result, nil
}

// UpdateDisposal updates a disposal with the given filter and update.
// Both the filter and update parameters should be MongoDB's BSON objects (primitives).
// It returns nil on success, and an error on failure.
func (ds *DatabaseService) UpdateDisposal(disposalToken string, update interface{}) error {
	r, err := ds.Client.Database(ds.dbName).Collection(DisposalCollectionName).UpdateOne(context.Background(),
		primitive.M{"token": disposalToken}, update)

	if err != nil {
		fmt.Printf("Failed to update disposal: %v\n", err)
		return err
	}

	if r.MatchedCount == 0 {
		fmt.Printf("No disposals matched the filter.\n")
		return structures.ErrNoDisposal
	}

	fmt.Printf("Updated disposal")

	return nil
}

func (ds *DatabaseService) LinkTransactionToUserById(transaction *structures.Transaction, userId string) error {
	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		fmt.Println("Invalid ID.")
		return structures.ErrInvalidDatabaseId
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{"$push": bson.M{"transactions": transaction}}

	res, err := ds.Client.Database(ds.dbName).Collection(UserCollectionName).UpdateOne(context.Background(),
		filter, update)

	if err != nil {
		fmt.Printf("Failed to link transaction to user %v: %v\n", userId, err)
		return err
	}

	if res.MatchedCount == 0 {
		fmt.Printf("No users matched the filter.\n")
		return structures.ErrNoUser
	}

	return nil
}
