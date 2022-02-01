package configs

import (
	Utils "Iagon/pkg/utils"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoInstance : MongoInstance Struct
type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// MI : An instance of MongoInstance Struct
var MI MongoInstance

func ConnectDB() {
	mongoUri, err := Utils.ConnectionURLBuilder("mongo")
	if err != nil {
		log.Fatal(err)
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database connected!")

	MI = MongoInstance{
		Client: client,
		DB:     client.Database(os.Getenv("DATABASE_NAME")),
	}
}

func ConnectionURLBuilder(s string) {
	panic("unimplemented")
}
