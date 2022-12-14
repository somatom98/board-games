package services

import (
	"context"
	"errors"
	"log"

	m "github.com/somatom98/board-games/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func FindMatch(id string) (m.IMatch, error) {
	matchesCollection, err := getCollection("games", "matches")
	if err != nil {
		return nil, err
	}

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	result := matchesCollection.FindOne(context.TODO(), bson.M{"_id": objId})
	var match m.IMatch
	if match, err = decodeToMatch(result); err != nil {
		return nil, err
	}
	return match, nil
}

func InsertMatch(match m.IMatch) error {
	matchesCollection, err := getCollection("games", "matches")
	if err != nil {
		return err
	}
	_, err = matchesCollection.InsertOne(context.TODO(), match)
	return err
}

func UpdateMatch(id primitive.ObjectID, board m.Board) error {
	matchesCollection, err := getCollection("games", "matches")
	if err != nil {
		return err
	}
	_, err = matchesCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.D{{"$set", bson.D{{"b", board}}}})
	return err
}

func FindGame(id primitive.ObjectID) (m.Game, error) {
	gameCollection, err := getCollection("games", "games")
	if err != nil {
		return m.Game{}, err
	}
	var game m.Game
	err = gameCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&game)
	return game, err
}

func FindGames() ([]m.Game, error) {
	gameCollection, err := getCollection("games", "games")
	if err != nil {
		return nil, err
	}
	cursor, err := gameCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	var games []m.Game
	if err = cursor.All(context.TODO(), &games); err != nil {
		log.Fatal(err)
	}
	return games, err
}

func getCollection(database string, collection string) (*mongo.Collection, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return client.Database(database).Collection(collection), nil
}

func decodeToMatch(result *mongo.SingleResult) (m.IMatch, error) {
	var bsonD bson.D
	err := result.Decode(&bsonD)
	if err != nil {
		return nil, err
	}
	game, err := FindGame(bsonD.Map()["g_id"].(primitive.ObjectID))
	if err != nil {
		return nil, err
	}
	switch game.Name {
	case "Quoridor":
		var match m.QuoridorMatch
		err = result.Decode(&match)
		return match, err
	default:
		return nil, errors.New("game id not existing")
	}
}
