package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *committeePublicKeyEventDB) lastBeaconProcess() uint64{
	opts := options.Find().SetSort(bson.D{{"BlkTimestamp", -1}}).SetLimit(1)

	cur, err := s.Collection.Find(context.TODO(), map[string]interface{}{
		"Event": "startepoch",
		"ChainID": -1,
	}, opts)

	if err != nil {
		fmt.Println(err)
		panic(1)
	}

	if !cur.Next(context.TODO()) {
		return 1
	}

	var tmp  bson.M
	err = cur.Decode(&tmp)
	if err != nil {
		panic(err)
	}
	return uint64(tmp["BlkHeight"].(int64))
}


func (s *committeePublicKeyEventDB) lastShardProcess(sid int) uint64{
	opts := options.Find().SetSort(bson.D{{"BlkTimestamp", -1}}).SetLimit(1)
	cur, err := s.Collection.Find(context.TODO(), bson.D{{"ChainID", sid},{"Event","startepoch"}	}, opts)

	if err != nil {
		fmt.Println(err)
		panic(1)
	}

	if !cur.Next(context.TODO()) {
		return 1
	}

	var tmp  bson.M
	err = cur.Decode(&tmp)
	if err != nil {
		panic(err)
	}
	return uint64(tmp["BlkHeight"].(int64))
}