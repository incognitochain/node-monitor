package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type committeePKInfo struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

func NewCommitteePKInfo(endpoint string, dbName, collectionName string) (*committeePKInfo, error) {
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
	keys := bson.D{{"CommitteePk", 1}, {"Epoch", 1}}
	index := mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	return &committeePKInfo{
		client, collection,
	}, nil
}

func (s *committeePKInfo) saveVoteStat(chainID int, epoch uint64, cpk string, totalVoteConfirm int, totalEpochCountBlock int) (err error) {

	filter := map[string]interface{}{
		"ChainID":     chainID,
		"CommitteePk": cpk,
		"Epoch":       int(epoch),
	}
	doc := bson.M{
		"$set": bson.M{
			"MiningPubkey":         getMiningPk(cpk),
			"totalVoteConfirm":     totalVoteConfirm,
			"totalEpochCountBlock": totalEpochCountBlock,
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

func (s *committeePKInfo) saveSlashing(epoch uint64, cpk string) (err error) {

	filter := map[string]interface{}{
		"CommitteePk": cpk,
		"Epoch":       int(epoch),
	}
	doc := bson.M{
		"$set": bson.M{
			"slashed": true,
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

func (s *committeePKInfo) saveReward(chainID int, epoch uint64, cpk string, rewardAmount int, time time.Time) (err error) {
	filter := map[string]interface{}{
		"ChainID":     chainID,
		"CommitteePk": cpk,
		"Epoch":       int(epoch),
	}
	doc := bson.M{
		"$set": bson.M{
			"Reward":       rewardAmount,
			"MiningPubkey": getMiningPk(cpk),
			"Time":         time,
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

func (s *committeePKInfo) saveBeaconReward(epoch uint64, cpk string, rewardAmount int, time time.Time) (err error) {

	filter := map[string]interface{}{
		"ChainID":     -1,
		"Epoch":       int(epoch),
		"CommitteePk": cpk,
	}
	doc := bson.M{
		"$set": bson.M{
			"MiningPubkey": getMiningPk(cpk),
			"Reward":       rewardAmount,
			"Time":         time,
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
