package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoClient struct {
	*mongo.Client
}


type committeePublicKeyEventDB struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}


func NewCommitteePublicKeyEventDB(endpoint string, dbName,collectionName string ) (*committeePublicKeyEventDB, error) {
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
	keys := bson.D{{"Event", -1}, {"CommitteePubkey", 1}, {"BlkTimestamp", 1}}
	index := mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"Event", -1}, {"RequestTx", 1}, {"BlkTimestamp", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"Event", -1},  {"BlkTimestamp", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"Event", -1}, {"ChainID", 1}, {"BlkTimestamp", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"Event", -1}, {"Epoch", 1}, {"ShardID", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"Event", -1}, {"Epoch", 1}, {"ChainID", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"CommitteePubkey", 1}, {"BlkTimestamp", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"CommitteePubkey", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"BlkTimestamp", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	opts = options.CreateIndexes().SetMaxTime(10 * time.Second)
	keys = bson.D{{"RewardReceiver", 1}}
	index = mongo.IndexModel{}
	index.Keys = keys
	collection.Indexes().CreateOne(context.Background(), index, opts)

	return &committeePublicKeyEventDB{
		client, collection,
	},	nil
}

func (s *committeePublicKeyEventDB) createStakeEvent(e StakeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "stake",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "stake",
		"FunderAddress" : e.FunderAddress,
		"CommitteePubkey": e.CommitteePubkey,
		"MiningPubkey": e.MiningPubkey,
		"RewardReceiver": e.RewardReceiver,
		"AutoStake": e.AutoStake,
		"ShardBlkHeight": e.ShardBlkHeight,
		"ShardBlkHash": e.ShardBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"Tx": e.Tx,
		"TxShardID": e.TxShardID,
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createStopAutoStakeEvent(e StopAutoStakeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "stopautostake",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "stopautostake",
		"CommitteePubkey": e.CommitteePubkey,
		"ShardBlkHeight": e.ShardBlkHeight,
		"ShardBlkHash": e.ShardBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"Tx": e.Tx,
		"TxShardID": e.TxShardID,
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createEpochRewardv3(e EpochReward) (err error){
	filter := map[string]interface{}{
		"Event": "reward",
		"Epoch": e.Epoch,
		"ShardID": e.ShardID,
	}
	doc := map[string]interface{}{
		"Event": "reward",
		"Epoch": e.Epoch,
		"ShardID": e.ShardID,
		"PRVAmountSet0": e.PRVAmountSet0,
		"PRVAmountSet1": e.PRVAmountSet1,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"BeaconBlkHash": e.BeaconBlkHash,

	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createEpochReward(e EpochReward) (err error){
	filter := map[string]interface{}{
		"Event": "reward",
		"Epoch": e.Epoch,
		"ShardID": e.ShardID,
	}
	doc := map[string]interface{}{
		"Event": "reward",
		"Epoch": e.Epoch,
		"ShardID": e.ShardID,
		"PRVAmount": e.PRVAmount,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"BeaconBlkHash": e.BeaconBlkHash,

	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createConfirmStakeEvent(e ConfirmStakeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "confirmstake",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "confirmstake",
		"CommitteePubkey": e.CommitteePubkey,
		"MiningPubKey": e.MiningPubKey,
		"RewardReceiver": e.RewardReceiver,
		"AutoStake": e.AutoStake,
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"BeaconBlkHash": e.BeaconBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createConfirmStopAutoStakeEvent(e ConfirmStopAutoStakeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "confirmstopautostake",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "confirmstopautostake",
		"CommitteePubkey": e.CommitteePubkey,
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"BeaconBlkHash": e.BeaconBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createAssignEvent(e AssignEvent) (err error){
	filter := map[string]interface{}{
		"Event": "assign",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "assign",
		"CommitteePubkey": e.CommitteePubkey,
		"AssignChainID": e.AssignChainID,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"BeaconBlkHash": e.BeaconBlkHash,
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createReturnStakingEvent(e ReturnStakingEvent) (err error){
	filter := map[string]interface{}{
		"Event": "returnstaking",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "returnstaking",
		"CommitteePubkey": e.CommitteePubkey,
		"ChainID": e.ChainID,
		"FunderAddress": e.FunderAddress,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"ReturnAmount": float64(e.ReturnAmount)*1e-9,
		"ReturnTx": e.ReturnTx,
		"ShardBlkHash": e.ShardBlkHash,
		"ShardBlkHeight": e.ShardBlkHeight,
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createConfirmJoinCommitteeEvent(e ConfirmJoinCommitteeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "confirmjoincommittee",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "confirmjoincommittee",
		"CommitteePubkey": e.CommitteePubkey,
		"JoinChainID": e.JoinChainID,
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"BeaconBlkHash": e.BeaconBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createConfirmLeaveCommitteeEvent(e ConfirmLeaveCommitteeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "confirmleavecommittee",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "confirmleavecommittee",
		"CommitteePubkey": e.CommitteePubkey,
		"BeaconBlkHeight": e.BeaconBlkHeight,
		"LeaveChainID": e.LeaveChainID,
		"BeaconBlkHash": e.BeaconBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createJoinCommitteeEvent(e JoinCommitteeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "joincommittee",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "joincommittee",
		"CommitteePubkey": e.CommitteePubkey,
		"JoinChainID": e.JoinChainID,
		"ShardBlkHeight": e.ShardBlkHeight,
		"ShardBlkHash": e.ShardBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createLeaveCommitteeEvent(e LeaveCommitteeEvent) (err error){
	filter := map[string]interface{}{
		"Event": "leavecommittee",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"CommitteePubkey": e.CommitteePubkey,
	}
	doc := map[string]interface{}{
		"Event": "leavecommittee",
		"CommitteePubkey": e.CommitteePubkey,
		"ShardBlkHeight": e.ShardBlkHeight,
		"LeaveChainID": e.LeaveChainID,
		"ShardBlkHash": e.ShardBlkHash,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}
	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createWithdrawRewardEvent(e WithdrawRewardEvent) (err error){
	filter := map[string]interface{}{
		"Event": "withdrawreward",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"RequestTx": e.RequestTx,
	}

	doc := map[string]interface{}{
		"Event": "withdrawreward",
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
		"RewardReceiver": e.RewardReceiver,
		"Amount": e.Amount,
		"Token": e.Token,
		"RequestTx": e.RequestTx,
		"ResponseTx": e.ResponseTx,
		"ChainID": e.ChainID,
		"ShardBlkHash": e.ShardBlkHash,
		"ShardBlkHeight": e.ShardBlkHeight,
	}

	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *committeePublicKeyEventDB) createStartEpoch(e StartEpoch) (err error){
	filter := map[string]interface{}{
		"Event": "startepoch",
		"ChainID": e.ChainID,
		"Epoch": e.Epoch,
	}

	doc := map[string]interface{}{
		"Event": "startepoch",
		"ChainID": e.ChainID,
		"Epoch": e.Epoch,
		"BlkHash": e.BlkHash,
		"BlkHeight": e.BlkHeight,
		"BlkTimestamp": time.Unix(int64(e.BlkTimestamp),0),
	}

	opts := options.FindOneAndReplace().SetUpsert(true)
	result := s.Collection.FindOneAndReplace(context.TODO(), filter,doc,opts)
	err = result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}
