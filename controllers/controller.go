package controllers

import (
	"advanced_2nd/models"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const connectionString = "mongodb+srv://rixy:rasul@cluster0.mxz6ddj.mongodb.net/?retryWrites=true&w=majority"
const dbName = "Animators"
const colName = "personInfo"

// MOST IMPORTANT
var collection *mongo.Collection

// connect with mongoDB
func init() {
	// client option
	clientOption := options.Client().ApplyURI(connectionString)

	// connect to mongoDB
	client, err := mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB connection success!")

	collection = client.Database(dbName).Collection(colName)

	// collection instance
	fmt.Println("Collection instance is ready!")
}

func insertOnePerson(person models.Person) {
	inserted, err := collection.InsertOne(context.Background(), person)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("inserted one Person: with id:", inserted.InsertedID)
}

func updateOnePerson(personId string) {
	id, _ := primitive.ObjectIDFromHex(personId)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"isChecked": true}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Checked Person:", result.ModifiedCount)
}
