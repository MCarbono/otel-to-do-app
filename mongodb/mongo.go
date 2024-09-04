package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func NewMongoDB(uri string) (*mongo.Client, error) {
	opts := options.Client()
	opts.Monitor = otelmongo.NewMonitor(otelmongo.WithCommandAttributeDisabled(false))
	opts.ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed connecting: %w", err)
	}
	docs := []interface{}{
		TodoModel{Title: "Buy groceries"},
		TodoModel{Title: "Study"},
		TodoModel{Title: "Walk with the dog"},
	}
	fmt.Println("seeding database!")
	_, err = client.Database("todo").Collection("todos").InsertMany(context.Background(), docs)
	if err != nil {
		return nil, fmt.Errorf("failed insertmany: %w", err)
	}
	fmt.Println("seeding successfully!")
	return client, nil
}

type TodoModel struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Title string             `bson:"title"`
}
