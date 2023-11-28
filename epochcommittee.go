package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type epochCommitteeDB struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

func NewEpochCommitteeDB(endpoint string, dbName, collectionName string) (*epochCommitteeDB, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(endpoint))
	if err != nil {
		panic("Cannot new client")
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		panic("Cannot connect")
		return nil, err
	}
	collection := client.Database(dbName).Collection(collectionName)

	//set indexing
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys := bson.D{{"ChainID", 1}, {"Height", 1}, {"Epoch", 1}}
	index := mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"ChainID", 1},  {"Epoch", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	return &epochCommitteeDB{
		client, collection,
	}, nil
}

func (s *epochCommitteeDB) saveCommitteeOnly(chainID int,  epoch uint64, commmittees []string) (err error) {

	filter := map[string]interface{}{
		"ChainID": chainID,
		"Epoch":   epoch,
	}
	doc := bson.M {
		"$set": bson.M {
			"Committee":      commmittees,
			"Size":           len(commmittees),
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	result := s.Collection.FindOneAndUpdate(context.TODO(), filter, doc, opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *epochCommitteeDB) saveCommittee(chainID int, height uint64, epoch uint64, commmittees []string, blkTime uint64, blkHash string) (err error) {

	filter := map[string]interface{}{
		"ChainID": chainID,
		"Epoch":   epoch,
	}
	doc := bson.M {
		"$set": bson.M {
			"Height":  height,
			"Committee":      commmittees,
			"Size":           len(commmittees),
			"BlockTime": time.Unix(int64(blkTime), 0),
			"BlockHash": blkHash,
			"From": 1,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	result := s.Collection.FindOneAndUpdate(context.TODO(), filter, doc, opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}
