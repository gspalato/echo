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

const USER_COLLECTION_NAME = "users"
const DISPOSAL_COLLECTION_NAME = "disposals"

type DatabaseService struct {
	client *mongo.Client

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

	ds.client = client

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

	err = ds.client.Database(ds.dbName).Collection(USER_COLLECTION_NAME).FindOne(
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

	err := ds.client.Database(ds.dbName).Collection(USER_COLLECTION_NAME).FindOne(
		context.Background(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, structures.ErrNoUser
	} else if err != nil {
		fmt.Printf("Failed to get user %v: %v\n", username, err)
		return nil, err
	}

	return &result, nil
}

func (ds *DatabaseService) GetDisposalsByUserId(userId string) ([]structures.DisposalClaim, error) {
	var result *[]structures.DisposalClaim

	filter := bson.M{"user_id": userId}

	cur, err := ds.client.Database(ds.dbName).Collection(DISPOSAL_COLLECTION_NAME).Find(
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

func (ds *DatabaseService) CreateUser(user structures.User) error {
	_, err := ds.client.Database(ds.dbName).Collection(USER_COLLECTION_NAME).InsertOne(context.Background(), user)
	if err != nil {
		fmt.Printf("Failed to create user %v: %v\n", user.Username, err)
		return err
	}

	fmt.Printf("Created user %v.\n", user.Username)

	return nil
}
