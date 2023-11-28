package main

//beacon block instruction
type ConfirmStakeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	MiningPubKey string
	RewardReceiver string
	AutoStake bool
	BeaconBlkHeight uint64
	BeaconBlkHash string
}

//beacon block instruction
type EpochReward struct {
	BlkTimestamp uint64
	Epoch uint64
	BeaconBlkHeight uint64
	BeaconBlkHash string
	ShardID int

	//ver1 ver2
	PRVAmount uint64

	//ver3
	PRVAmountSet0 uint64
	PRVAmountSet1 uint64

}


//shard block action
type StakeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	FunderAddress string
	CommitteePubkey string
	MiningPubkey string
	RewardReceiver string
	AutoStake bool
	TxShardID int
	Tx string
	ShardBlkHeight uint64
	ShardBlkHash string
}

//beacon block instruction
type ConfirmStopAutoStakeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	BeaconBlkHeight uint64
	BeaconBlkHash string
}

//shard block action
type StopAutoStakeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	TxShardID int
	Tx string
	ShardBlkHeight uint64
	ShardBlkHash string
}

//beacon block instruction
type AssignEvent struct {
	BlkTimestamp    uint64
	Epoch uint64
	CommitteePubkey string
	AssignChainID   int
	BeaconBlkHeight uint64
	BeaconBlkHash   string
}

//shard block action
type JoinCommitteeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	JoinChainID int
	ShardBlkHeight uint64
	ShardBlkHash string

}

//shard block action
type LeaveCommitteeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	LeaveChainID int
	ShardBlkHeight uint64
	ShardBlkHash string
}

//beacon block instruction
type ConfirmJoinCommitteeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	JoinChainID int
	BeaconBlkHeight uint64
	BeaconBlkHash string

}

//beacon block instruction
type ConfirmLeaveCommitteeEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	LeaveChainID int
	BeaconBlkHeight uint64
	BeaconBlkHash string
}

//shard block action
type ReturnStakingEvent struct {
	BlkTimestamp uint64
	Epoch uint64
	CommitteePubkey string
	FunderAddress string
	ShardBlkHeight uint64
	ShardBlkHash string
	ChainID int
	ReturnAmount int
	ReturnTx string
}

type WithdrawRewardEvent struct {
	BlkTimestamp    uint64
	Epoch uint64
	RewardReceiver  string
	Token string
	Amount          uint64
	RequestTx       string
	ResponseTx      string
	ShardBlkHeight uint64
	ShardBlkHash string
	ChainID int
}

type StartEpoch struct {
	BlkTimestamp    uint64
	BlkHeight uint64
	Epoch uint64
	BlkHash string
	ChainID int
}

//v2 force join/leave event from beacon