package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type statDB struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

func NewStatDB(endpoint string, dbName, collectionName string) (*statDB, error) {
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
	keys := bson.D{{"ChainID", -1}, {"BlockHeight", 1}}
	index := mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	return &statDB{
		client, collection,
	}, nil
}

func (s *statDB) lastBlock(cid int) uint64 {
	opts := options.Find().SetSort(bson.D{{"BlockHeight", -1}}).SetLimit(1)

	cur, err := s.Collection.Find(context.TODO(), map[string]interface{}{
		"ChainID": cid,
	}, opts)

	if err != nil {
		fmt.Println(err)
		panic(1)
	}

	if !cur.Next(context.TODO()) {
		return 1
	}

	var tmp bson.M
	err = cur.Decode(&tmp)
	if err != nil {
		panic(err)
	}
	return uint64(tmp["BlockHeight"].(int32))
}

type StatInfo struct {
	ChainID       int
	BlockHeight   int
	NumInst       int
	NumShardState int
	NumTx         int
	NumCrossTx    int
	BlockHash     string
	BlockTime     time.Time
}

func (s *statDB) set(info StatInfo) error {
	filter := map[string]interface{}{
		"ChainID":     info.ChainID,
		"BlockHeight": info.BlockHeight,
	}
	doc := bson.M{
		"$set": bson.M{
			"ChainID":       info.ChainID,
			"BlockHeight":   info.BlockHeight,
			"NumInst":       info.NumInst,
			"NumShardState": info.NumShardState,
			"NumTx":         info.NumTx,
			"NumCrossTx":    info.NumCrossTx,
			"BlockHash":     info.BlockHash,
			"BlockTime":     info.BlockTime,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	result := s.Collection.FindOneAndUpdate(context.TODO(), filter, doc, opts)
	err := result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}
