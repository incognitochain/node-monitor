package main

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"time"
)

func tx_stat(){
	statDB, err := NewStatDB(MONGODB, "netmonitor", "stat")
	if err != nil {
		panic(err)
	}
	node := NewAppService(INCOGNITO_NODE, true)
	fromBeaconBlock := statDB.lastBlock(-1)
	node.OnBeaconBlock(uint64(int64(fromBeaconBlock)),func (beaconBlk types.BeaconBlock) {
		numShardState := 0
		for _, v := range beaconBlk.Body.ShardState {
			numShardState += len(v)
		}
		//fmt.Println("beacon",beaconBlk.GetHeight())
		statDB.set(StatInfo{
			ChainID : -1,
			BlockHeight : int(beaconBlk.GetHeight()),
			NumInst : len(beaconBlk.Body.Instructions),
			NumShardState :numShardState,
			NumTx: 0,
			NumCrossTx:0,
			BlockHash: beaconBlk.Hash().String(),
			BlockTime: time.Unix(beaconBlk.GetProduceTime(),0),
		})
	})

	for i := 0; i < config.Param().ActiveShards; i++ {
		fromBlock := statDB.lastBlock(i)
		node.OnShardBlock(i, uint64(int64(fromBlock)),  func(shardBlk types.ShardBlock) {
			shardID := shardBlk.GetShardID()
			//fmt.Println("shardBlk",shardID, shardBlk.GetHeight())
			statDB.set(StatInfo{
				ChainID : shardID,
				BlockHeight : int(shardBlk.GetHeight()),
				NumInst : len(shardBlk.Body.Instructions),
				NumShardState :0,
				NumTx: len(shardBlk.Body.Transactions),
				NumCrossTx:len(shardBlk.Body.CrossTransactions),
				BlockHash: shardBlk.Hash().String(),
				BlockTime: time.Unix(shardBlk.GetProduceTime(),0),
			})
		})
	}
}