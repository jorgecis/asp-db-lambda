package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `json:"name"`
	Password      string             `json:"password"`
	Email         string             `json:"email"`
	CreatedAt     time.Time          `json:"createdAt"`
	Level         int                `json:"level"`
	Active        bool               `json:"active"`
	Id_Dealer     int                `json:"id_dealer"`
	DealerName    string             `json:"dealer_name"`
	Id_Insurace   int                `json:"id_insurance"`
	InsuranceName string             `json:"insurance_name"`
}

var mgdb *mongo.Client

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func init() {
	var err error
	dbUser := getenv("db_user", "aspuser")
	dbPass := getenv("db_pass", "")
	dbHost := getenv("db_host", "172.31.17.185:27017")
	dbName := getenv("db_name", "asp")

	log.Println("Connecting to DB...")
	dsn := fmt.Sprintf("mongodb://%s:%s@%s/%s",
		url.QueryEscape(dbUser),
		url.QueryEscape(dbPass),
		dbHost,
		url.QueryEscape(dbName),
	)

	ctx := context.Background()
	mgdb, err = mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		log.Fatal(err)
	}
	if err = mgdb.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to DB")
}

// Validate_email checks that the sender is an authorized user in MongoDB.
func Validate_email(email string) bool {
	collection := mgdb.Database("asp").Collection("users")

	filter := bson.M{
		"email":  email,
		"active": true,
		"level":  10,
	}
	var user User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
