package main

import (
	"context"
	"encoding/json"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/incognitochain/incognito-chain/wallet"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MONGODB = "mongodb://admin:adminzzz@localhost:27017/admin"

const INCOGNITO_NODE = "http://<fullnode-ip>:<rpc-port>"

var cacheMpk = make(map[string]string)
var committeeFromBlockEpoch, _ = lru.New(1000)

func getMiningPk(cpk string) string {
	if cacheMpk[cpk] != "" {
		return cacheMpk[cpk]
	}
	MiningPubkey := incognitokey.CommitteePublicKey{}
	MiningPubkey.FromString(cpk)
	mpk := MiningPubkey.GetMiningKeyBase58(common.BlsConsensus)
	cacheMpk[cpk] = mpk
	return mpk
}

func main() {

	config.LoadConfig()
	config.LoadParam()
	portal.SetupParam()
	remoteRPC := RemoteRPCClient{INCOGNITO_NODE}

	lastEpoch := sync.Map{}
	//node := devframework.NewAppNode("fullnode", devframework.MainNetParam, false, true)
	//node.StopSync()
	node := NewAppService(INCOGNITO_NODE, true)

	eventDB, err := NewCommitteePublicKeyEventDB(MONGODB, "netmonitor", "committeeEvent")
	committeeDB, err := NewEpochCommitteeDB(MONGODB, "netmonitor", "epochCommittee")
	committeePkInfoDB, err := NewCommitteePKInfo(MONGODB, "netmonitor", "committeePkInfo")

	if err != nil {
		panic(err)
	}

	fromBeaconBlk := eventDB.lastBeaconProcess()
	epochCommitteeVote := make(map[int]map[string]int) //sid -> cpk -> vote stat
	shardCommittee := make(map[int][]string)
	totalEpochConfirm := make(map[int]int)
	totalEpochConfirmV3 := make(map[int][]int)
	lastShardStateValidatorCommittee := map[int][]string{}
	lastShardStateValidatorIndex := map[int][]int{}

	startEpoch := uint64(0)
	beaconCommitteeState, err := remoteRPC.GetCommitteeStateOnBeacon(fromBeaconBlk, "")
	if err != nil {
		fmt.Println(err)
		panic("Cannot get committee state on beacon")
	}
	shardCommittee = beaconCommitteeState.Committee

	for sid := 0; sid < 8; sid++ {
		epochCommitteeVote[int(sid)] = make(map[string]int)
		totalEpochConfirm[sid] = 0
		totalEpochConfirmV3[sid] = []int{0, 0}
	}

	node.OnBeaconBlock(uint64(int64(fromBeaconBlk)), func(beaconBlk types.BeaconBlock) {

		blkHash := beaconBlk.Hash()
		blkHeight := beaconBlk.GetHeight()
		if blkHeight == 1917*350+1 {
			beaconCommitteeState, err = remoteRPC.GetCommitteeStateOnBeacon(blkHeight, "")
			if err != nil {
				fmt.Println(err)
				panic("Cannot get committee state on beacon")
			}
			shardCommittee = beaconCommitteeState.Committee
		}
		//fmt.Println(beaconBlk.Body.ShardState)
		if startEpoch == 0 {
			startEpoch = beaconBlk.GetCurrentEpoch()
		}

		shardSwapoutCommittee := make(map[int][]string)

	REPEAT:
		for _, inst := range beaconBlk.Body.Instructions {
			switch inst[0] {
			case "39":
				reward := &metadata.BeaconRewardInfo{}
				err := json.Unmarshal([]byte(inst[3]), reward)
				if err != nil {
					panic(err)
				}
				prvReward := reward.BeaconReward[common.PRVCoinID]
				publicKey := reward.PayToPublicKey
				beaconCommittee := shardCommittee[-1]
				for _, cpk := range beaconCommittee {
					acc, _ := account.NewAccountFromCommitteePublicKey(cpk)
					if acc.PublicKey == publicKey {
						//fmt.Println(beaconBlk.GetCurrentEpoch()-1, cpk, getMiningPk(cpk), int(prvReward), time.Unix(beaconBlk.GetProduceTime(), 0))
						err = committeePkInfoDB.saveBeaconReward(beaconBlk.GetCurrentEpoch()-1, cpk, int(prvReward), time.Unix(beaconBlk.GetProduceTime(), 0))
						if err != nil && err.Error() == "mongo: no documents in result" {
							panic(err)
						}
					}
				}
			case "shardreceiverewardv3":
				rewardInst, _ := instruction.ValidateAndImportShardReceiveRewardV3InstructionFromString(inst)
				//fmt.Println("shardreceiverewardv3", rewardInst)
				rewardAmount0 := uint64(0)
				rewardAmount1 := uint64(0)
				rewardAmount := uint64(0)
				if rewardInst.SubsetID() == 0 {
					rewardAmount0 = rewardInst.Reward()[common.PRVCoinID]
					rewardAmount = rewardAmount0
				} else {
					rewardAmount1 = rewardInst.Reward()[common.PRVCoinID]
					rewardAmount = rewardAmount1
				}
				err = eventDB.createEpochReward(EpochReward{
					uint64(beaconBlk.GetProduceTime()),
					uint64(rewardInst.Epoch()),
					blkHeight,
					blkHash.String(),
					int(rewardInst.ShardID()),
					0, rewardAmount0, rewardAmount1,
				})
				if err != nil {
					panic(err)
				}
				query := bson.M{"Epoch": rewardInst.Epoch(), "ChainID": rewardInst.ShardID()}
				singleRes := committeeDB.Collection.FindOne(context.TODO(), query)
				var tmp = bson.M{}
				singleRes.Decode(&tmp)
				if size, ok := tmp["Size"]; !ok {
					fmt.Println(bson.M{"Epoch": rewardInst.Epoch(), "ChainID": rewardInst.ShardID()})
					time.Sleep(time.Second)
					goto REPEAT
				} else {
					numSubsetCandidate := 0
					for i := int32(0); i < size.(int32); i++ {
						if i%2 == int32(rewardInst.SubsetID()) {
							numSubsetCandidate++
						}
					}
					eachCandidateAmount := rewardAmount / uint64(numSubsetCandidate)
					//fmt.Println(eachCandidateAmount)
					if rewardInst.SubsetID() == 0 {
						doc := bson.M{
							"$set": bson.M{
								"RewardAmount":  eachCandidateAmount,
								"RewardAmount0": eachCandidateAmount,
							},
						}
						opts := options.FindOneAndUpdate().SetUpsert(false)
						result := committeeDB.Collection.FindOneAndUpdate(context.TODO(), query, doc, opts)
						err = result.Err()
						if err != nil && err.Error() == "mongo: no documents in result" {
							panic(err)
						}
					} else {
						doc := bson.M{
							"$set": bson.M{
								"RewardAmount1": eachCandidateAmount,
							},
						}
						opts := options.FindOneAndUpdate().SetUpsert(false)
						result := committeeDB.Collection.FindOneAndUpdate(context.TODO(), query, doc, opts)
						err = result.Err()
						if err != nil && err.Error() == "mongo: no documents in result" {
							panic(err)
						}
					}

					committee := tmp["Committee"].(bson.A)
					for i, cpk := range committee {
						if i%2 == int(rewardInst.SubsetID()) {
							err = committeePkInfoDB.saveReward(int(rewardInst.ShardID()), uint64(rewardInst.Epoch()), cpk.(string), int(eachCandidateAmount), time.Unix(beaconBlk.GetProduceTime(), 0))
							//fmt.Println(cpk.(string), uint64(rewardInst.Epoch()))
							if err != nil && err.Error() == "mongo: no documents in result" {
								panic(err)
							}
						}
					}
				}
			case "43":
				shardID := inst[1]
				sID, _ := strconv.Atoi(shardID)
				type RewardInst struct {
					ShardReward map[string]uint64
					Epoch       int
				}
				var reInst RewardInst
				err := json.Unmarshal([]byte(inst[3]), &reInst)
				if err != nil {
					panic(err)
				}
				err = eventDB.createEpochReward(EpochReward{
					uint64(beaconBlk.GetProduceTime()),
					uint64(reInst.Epoch),
					blkHeight,
					blkHash.String(),
					sID,
					reInst.ShardReward["0000000000000000000000000000000000000000000000000000000000000004"],
					0, 0,
				})
				if err != nil {
					panic(err)
				}

				query := bson.M{"Epoch": reInst.Epoch, "ChainID": sID}
				singleRes := committeeDB.Collection.FindOne(context.TODO(), query)
				var tmp = bson.M{}
				singleRes.Decode(&tmp)
				if size, ok := tmp["Size"]; !ok {
					fmt.Println(bson.M{"Epoch": reInst.Epoch, "ChainID": sID})
					time.Sleep(time.Second)
					goto REPEAT
				} else {
					eachCandidateAmount := reInst.ShardReward["0000000000000000000000000000000000000000000000000000000000000004"] / uint64(size.(int32))
					doc := bson.M{
						"$set": bson.M{
							"RewardAmount": eachCandidateAmount,
						},
					}
					opts := options.FindOneAndUpdate().SetUpsert(false)
					result := committeeDB.Collection.FindOneAndUpdate(context.TODO(), query, doc, opts)
					err = result.Err()
					if err != nil && err.Error() == "mongo: no documents in result" {
						panic(err)
					}

					committee := tmp["Committee"].(bson.A)
					for _, cpk := range committee {
						err = committeePkInfoDB.saveReward(sID, uint64(reInst.Epoch), cpk.(string), int(eachCandidateAmount), time.Unix(beaconBlk.GetProduceTime(), 0))
						if err != nil && err.Error() == "mongo: no documents in result" {
							panic(err)
						}
					}
				}
			case "stake":
				committeePk := strings.Split(inst[1], ",")
				//stakeChain := inst[2]
				//shardID := inst[3]
				rewardReceiver := strings.Split(inst[4], ",")
				autoStaking := strings.Split(inst[5], ",")

				for i, _ := range committeePk {
					var autoStake bool
					if autoStaking[i] == "true" {
						autoStake = true
					} else if autoStaking[i] == "false" {
						autoStake = false
					} else {
						panic("something wrong here!")
					}

					MiningPubkey := incognitokey.CommitteePublicKey{}
					MiningPubkey.FromString(committeePk[i])
					mpk := MiningPubkey.GetMiningKeyBase58(common.BlsConsensus)

					err := eventDB.createConfirmStakeEvent(ConfirmStakeEvent{
						uint64(beaconBlk.GetProduceTime()),
						beaconBlk.GetCurrentEpoch(),
						committeePk[i],
						mpk,
						rewardReceiver[i],
						autoStake,
						blkHeight,
						blkHash.String(),
					})
					if err != nil {
						panic(err)
					}

				}
			case "stopautostake":
				stopPks := strings.Split(inst[1], ",")
				for i, _ := range stopPks {
					err := eventDB.createConfirmStopAutoStakeEvent(ConfirmStopAutoStakeEvent{
						uint64(beaconBlk.GetProduceTime()),
						beaconBlk.GetCurrentEpoch(),
						stopPks[i],
						blkHeight,
						blkHash.String(),
					})
					if err != nil {
						panic(err)
					}

				}
			case "assign":
				assignPks := strings.Split(inst[1], ",")
				shardID := inst[3]
				sID, _ := strconv.Atoi(shardID)
				for i, _ := range assignPks {
					err := eventDB.createAssignEvent(AssignEvent{
						uint64(beaconBlk.GetProduceTime()),
						beaconBlk.GetCurrentEpoch(),
						assignPks[i],
						sID,
						blkHeight,
						blkHash.String(),
					})
					if err != nil {
						panic(err)
					}

				}
			case "random":
				//donothing
			case instruction.SWAP_ACTION, instruction.SWAP_SHARD_ACTION:
				joinCommittee := strings.Split(inst[1], ",")
				leaveCommittee := strings.Split(inst[2], ",")
				if inst[1] == "" {
					joinCommittee = []string{}
				}
				if inst[2] == "" {
					leaveCommittee = []string{}
				}
				sID := -2
				if inst[0] == instruction.SWAP_ACTION {
					sID, _ = strconv.Atoi(inst[4])
				}
				if inst[0] == instruction.SWAP_SHARD_ACTION {
					sID, _ = strconv.Atoi(inst[3])
					shardSwapoutCommittee[sID] = leaveCommittee //mark cpk is swap out committee
				}

				if sID == -2 {
					panic(inst)
				}
				for i, _ := range joinCommittee {
					if joinCommittee[i] == "" {
						continue
					}
					err := eventDB.createConfirmJoinCommitteeEvent(ConfirmJoinCommitteeEvent{
						uint64(beaconBlk.GetProduceTime()),
						beaconBlk.GetCurrentEpoch(),
						joinCommittee[i],
						sID,
						blkHeight,
						blkHash.String(),
					})
					if err != nil {
						panic(err)
					}

				}

				for i, _ := range leaveCommittee {
					if leaveCommittee[i] == "" {
						continue
					}

					err := eventDB.createConfirmLeaveCommitteeEvent(ConfirmLeaveCommitteeEvent{
						uint64(beaconBlk.GetProduceTime()),
						beaconBlk.GetCurrentEpoch(),
						leaveCommittee[i],
						sID,
						blkHeight,
						blkHash.String(),
					})
					if err != nil {
						panic(err)
					}
				}
			}
		}

		beaconLastEpoch, ok := lastEpoch.Load(-1)
		committeeFromBlockEpoch.Add(beaconBlk.Hash().String(), beaconBlk.GetCurrentEpoch())
		if beaconBlk.GetHeight() >= config.Param().ConsensusParam.StakingFlowV2Height {
			//save committee vote state for realtime block
			if time.Since(time.Unix(beaconBlk.GetProduceTime(), 0)).Seconds() < 120 {
				for sid, validators := range shardCommittee {
					for index, cpk := range validators {
						if epochCommitteeVote[sid] != nil {
							if beaconBlk.GetVersion() == 7 || beaconBlk.GetVersion() == 8 {
								committeePkInfoDB.saveVoteStat(sid, beaconLastEpoch.(uint64), cpk, epochCommitteeVote[sid][cpk], totalEpochConfirmV3[int(sid)][index%2])
							} else {
								committeePkInfoDB.saveVoteStat(sid, beaconLastEpoch.(uint64), cpk, epochCommitteeVote[sid][cpk], totalEpochConfirm[int(sid)])
							}

						}
					}
				}
			}
		}

		if (ok && beaconBlk.GetCurrentEpoch() != beaconLastEpoch) || (!ok && beaconBlk.GetCurrentEpoch() == 1) {
			fmt.Println("process new epoch", beaconBlk.GetCurrentEpoch(), totalEpochConfirm)
			if beaconBlk.GetHeight() >= config.Param().ConsensusParam.StakingFlowV2Height {
				//save committee vote state in last epoch
				for sid, validators := range shardCommittee {
					for index, cpk := range validators {
						if epochCommitteeVote[sid] != nil {
							if beaconBlk.GetVersion() == 7 || beaconBlk.GetVersion() == 8 {
								if totalEpochConfirmV3[int(sid)][index%2] < epochCommitteeVote[sid][cpk] {
									epochCommitteeVote[sid][cpk] = totalEpochConfirmV3[int(sid)][index%2]
								}
								committeePkInfoDB.saveVoteStat(sid, beaconLastEpoch.(uint64), cpk, epochCommitteeVote[sid][cpk], totalEpochConfirmV3[int(sid)][index%2])
							} else {
								committeePkInfoDB.saveVoteStat(sid, beaconLastEpoch.(uint64), cpk, epochCommitteeVote[sid][cpk], totalEpochConfirm[int(sid)])
							}
						}
					}
				}

				//get cpk that is exit of chain (assume if not exit, will go to subtitute)
				beaconCommitteeState, err := remoteRPC.GetCommitteeStateOnBeacon(blkHeight, "")
				if err != nil {
					fmt.Println(err)
					panic("Cannot get committee state on beacon")
				}

				//check autostaking for kickout cpk (exit and not in Substitute), if still autostake but kick then it is slashed
				bView, err := remoteRPC.GetBeaconViewByHash(beaconBlk.GetPrevHash().String())
				if len(bView.ShardCommittee) != 8 {
					panic("Something wrong with get beacon view by hash" + beaconBlk.GetPrevHash().String())
				}
				for sid, cpks := range bView.ShardCommittee { //loop previous cpk committee
					//fmt.Println("cpks", sid, len(cpks))
					for _, cpk := range cpks {
						exit := common.IndexOfStr(cpk, beaconCommitteeState.Committee[int(sid)]) == -1
						if exit {
							kickOut := true
							for _, subtitutes := range beaconCommitteeState.Substitute {
								if common.IndexOfStr(cpk, subtitutes) > -1 {
									kickOut = false
									break
								}
							}
							//fmt.Println("cpk", cpk, kickOut)
							if kickOut {
								autostaking, ok := bView.AutoStaking[cpk]
								if !ok {
									//fmt.Println("cpk", cpk)
									panic(fmt.Sprintf("something wrong with %v has no autostaking value %v %V", cpk, beaconBlk.GetPrevHash().String()))
								}
								//fmt.Println("autostaking", autostaking)
								if autostaking {
									committeePkInfoDB.saveSlashing(beaconLastEpoch.(uint64), cpk)
								}
							}
						}

					}
				}

				//update committee for new epoch
				shardCommittee = beaconCommitteeState.Committee
				for sid := 0; sid < 8; sid++ {
					epochCommitteeVote[int(sid)] = make(map[string]int)
					totalEpochConfirm[sid] = 0
					totalEpochConfirmV3[sid] = []int{0, 0}
					err = committeeDB.saveCommittee(sid, blkHeight, beaconBlk.Header.Epoch, shardCommittee[sid], uint64(beaconBlk.GetProduceTime()), beaconBlk.Hash().String())
					if err != nil {
						panic(err)
					}
				}
				//panic(1)
			}

			startEpoch = beaconBlk.GetCurrentEpoch()

			err = eventDB.createStartEpoch(StartEpoch{
				uint64(beaconBlk.GetProduceTime()),
				beaconBlk.GetHeight(),
				beaconBlk.GetCurrentEpoch(),
				beaconBlk.Hash().String(),
				-1,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}

		}
		lastEpoch.Store(-1, beaconBlk.GetCurrentEpoch())

		//update vote for current committee
		if beaconBlk.GetHeight() >= config.Param().ConsensusParam.StakingFlowV2Height {
			for sid, shardState := range beaconBlk.Body.ShardState {
				for _, blkShard := range shardState {
					subset := int(blkShard.ProposerTime/40) % 2
					valData, err := consensustypes.DecodeValidationData(blkShard.ValidationData)
					if err != nil {
						panic(err)
						return
					}

					//check if this shard block is new epoch ~ new committee -> reset lastShardStateValdiator
					blkShardEpoch, _ := committeeFromBlockEpoch.Get(blkShard.CommitteeFromBlock.String())
					//fmt.Println("epoch check:", beaconBlk.GetHeight(), blkShardEpoch, startEpoch)
					if blkShardEpoch != startEpoch {
						if blkShard.Version >= 8 {
							lastShardStateValidatorCommittee[int(sid)] = []string{}
							lastShardStateValidatorIndex[int(sid)] = []int{}
						}
						continue
					}

					totalEpochConfirm[int(sid)]++
					if blkShard.Version == 7 || blkShard.Version == 8 {
						totalEpochConfirmV3[int(sid)][subset]++
					}

					if blkShard.PreviousValidationData != "" {
						prevValidationData, err := consensustypes.DecodeValidationData(blkShard.PreviousValidationData)
						if err != nil {
							panic(err)
							return
						}
						tempToBeSignedCommittees := lastShardStateValidatorCommittee[int(sid)]
						prevShardStateValidatorIndex := lastShardStateValidatorIndex[int(sid)]
						uncountCommittees := make(map[string]struct{})
						for _, idx := range prevValidationData.ValidatiorsIdx {
							if idx >= len(tempToBeSignedCommittees) {
								if len(tempToBeSignedCommittees) == 0 {
									break
								}
								fmt.Println(beaconBlk.GetHeight(), sid, blkShard.Height, blkShard.Hash.String(), prevValidationData.ValidatiorsIdx, len(tempToBeSignedCommittees))
								panic("something wrong with previous validation data")
							}
							if common.IndexOfInt(idx, prevShardStateValidatorIndex) == -1 {
								uncountCommittees[tempToBeSignedCommittees[idx]] = struct{}{}
							}
						}

						for _, toBeSignedCommittee := range tempToBeSignedCommittees {
							if _, ok := uncountCommittees[toBeSignedCommittee]; ok {
								if _, ok := epochCommitteeVote[int(sid)][toBeSignedCommittee]; ok {
									epochCommitteeVote[int(sid)][toBeSignedCommittee]++
								} else {
									epochCommitteeVote[int(sid)][toBeSignedCommittee] = 0
								}
							}
						}

					}
					//fmt.Println(sid,subset,valData.ValidatiorsIdx, len(shardCommittee[int(sid)]))
					//fmt.Println(sid, valData.ValidatiorsIdx, blkHash,blkShard.Hash,blkShard.CommitteeFromBlock)
					for _, i := range valData.ValidatiorsIdx {
						committeePK := shardCommittee[int(sid)][i]
						if blkShard.Version == 7 || blkShard.Version == 8 {
							if subset == 0 {
								committeePK = shardCommittee[int(sid)][i*2]
							} else {
								committeePK = shardCommittee[int(sid)][i*2+1]
							}
						}
						epochCommitteeVote[int(sid)][committeePK]++
					}

					if blkShard.Version >= 8 {
						lastShardStateValidatorCommittee[int(sid)] = []string{}
						for i, cpk := range shardCommittee[int(sid)] {
							if blkShard.Version == 8 {
								if i%2 == subset {
									lastShardStateValidatorCommittee[int(sid)] = append(lastShardStateValidatorCommittee[int(sid)], cpk)
								}
							} else {
								lastShardStateValidatorCommittee[int(sid)] = append(lastShardStateValidatorCommittee[int(sid)], cpk)
							}
						}
						lastShardStateValidatorIndex[int(sid)] = valData.ValidatiorsIdx
					}

				}
			}
		}

	})
	//

	for i := 0; i < config.Param().ActiveShards; i++ {
		fromBlock := eventDB.lastShardProcess(i)
		if fromBlock > 1 {
			fromBlock--
		}
		fmt.Println("last shard ", i, fromBlock)
		node.OnShardBlock(i, uint64(int64(fromBlock)), func(shardBlk types.ShardBlock) {
			blkHash := shardBlk.Hash()

			blkHeight := shardBlk.GetHeight()
			chainID := int(shardBlk.Header.ShardID)
			//fmt.Println("shard",chainID, blkHeight)
			for _, tx := range shardBlk.Body.Transactions {
				switch tx.GetMetadataType() {
				case metadata.WithDrawRewardResponseMeta:
					withdrawReward, ok := tx.GetMetadata().(*metadata.WithDrawRewardResponse)
					if !ok {
						panic("something wrong when parse withdrawReward tx")
					}
					_, _, amount, _ := tx.GetTransferData()
					txRequest := withdrawReward.TxRequest

					var txData metadata.Transaction
					for _, txReq := range shardBlk.Body.Transactions {
						if txRequest.String() == txReq.Hash().String() {
							txData = txReq
							break
						}
					}

					txRequestMeta, ok := txData.GetMetadata().(*metadata.WithDrawRewardRequest)

					if !ok {
						panic("WithDrawRewardRequest wrong")
					}

					keyWL := wallet.KeyWallet{}
					keyWL.KeySet.PaymentAddress = txRequestMeta.PaymentAddress
					payment := keyWL.Base58CheckSerialize(wallet.PaymentAddressType)
					eventDB.createWithdrawRewardEvent(WithdrawRewardEvent{
						BlkTimestamp:   uint64(shardBlk.GetProduceTime()),
						Epoch:          shardBlk.GetCurrentEpoch(),
						RewardReceiver: payment,
						Amount:         amount,
						ResponseTx:     tx.Hash().String(),
						RequestTx:      txData.Hash().String(),
						Token:          txRequestMeta.TokenID.String(),
						ShardBlkHeight: blkHeight,
						ShardBlkHash:   blkHash.String(),
						ChainID:        chainID,
					})

				case metadata.ReturnStakingMeta:
					returnstaking, ok := tx.GetMetadata().(*metadata.ReturnStakingMetadata)
					if !ok {
						panic("something wrong when parse return staking tx")
					}
					tx := returnstaking.TxID
					txHash, _ := common.Hash{}.NewHashFromStr(tx)
					txDetail, err := remoteRPC.GetTransactionByHash(tx)
					if err != nil {
						fmt.Println(returnstaking)
						fmt.Println(tx)
						fmt.Println(err)
						panic("process returnstaking error 1")
					}
					var stakingMetaData = new(metadata.StakingMetadata)
					err = json.Unmarshal([]byte(txDetail.Metadata), stakingMetaData)
					if err != nil {
						fmt.Println(err)
						fmt.Println(tx)
						fmt.Println(txDetail.Metadata)
						fmt.Println(stakingMetaData)
						panic("process returnstaking error 2")
					}
					eventDB.createReturnStakingEvent(ReturnStakingEvent{
						BlkTimestamp:    uint64(shardBlk.GetProduceTime()),
						Epoch:           shardBlk.GetCurrentEpoch(),
						CommitteePubkey: stakingMetaData.CommitteePublicKey,
						FunderAddress:   stakingMetaData.FunderPaymentAddress,
						ShardBlkHeight:  blkHeight,
						ShardBlkHash:    blkHash.String(),
						ChainID:         chainID,
						ReturnAmount:    int(stakingMetaData.StakingAmountShard),
						ReturnTx:        txHash.String(),
					})

				case metadata.ShardStakingMeta:
					stakingMetadata, ok := tx.GetMetadata().(*metadata.StakingMetadata)
					if !ok {
						panic("something wrong when parse stake tx")
					}
					MiningPubkey := incognitokey.CommitteePublicKey{}
					MiningPubkey.FromString(stakingMetadata.CommitteePublicKey)
					mpk := MiningPubkey.GetMiningKeyBase58(common.BlsConsensus)
					eventDB.createStakeEvent(StakeEvent{
						BlkTimestamp:    uint64(shardBlk.GetProduceTime()),
						Epoch:           shardBlk.GetCurrentEpoch(),
						FunderAddress:   stakingMetadata.FunderPaymentAddress,
						CommitteePubkey: stakingMetadata.CommitteePublicKey,
						MiningPubkey:    mpk,
						RewardReceiver:  stakingMetadata.RewardReceiverPaymentAddress,
						AutoStake:       stakingMetadata.AutoReStaking,
						TxShardID:       int(shardBlk.Header.ShardID),
						Tx:              tx.Hash().String(),
						ShardBlkHeight:  blkHeight,
						ShardBlkHash:    blkHash.String(),
					})
				case metadata.StopAutoStakingMeta:
					stopAutoStakingMetadata, ok := tx.GetMetadata().(*metadata.StopAutoStakingMetadata)
					if !ok {
						panic("something wrong when parse stop auto stake")
					}
					eventDB.createStopAutoStakeEvent(StopAutoStakeEvent{
						BlkTimestamp:    uint64(shardBlk.GetProduceTime()),
						Epoch:           shardBlk.GetCurrentEpoch(),
						CommitteePubkey: stopAutoStakingMetadata.CommitteePublicKey,
						TxShardID:       int(shardBlk.Header.ShardID),
						Tx:              tx.Hash().String(),
						ShardBlkHeight:  blkHeight,
						ShardBlkHash:    blkHash.String(),
					})
				}
			}

			joinCommittee := []string{}
			leaveCommittee := []string{}
			for _, inst := range shardBlk.Body.Instructions {
				switch inst[0] {
				case "swap":

					joinCommittee = strings.Split(inst[1], ",")
					leaveCommittee = strings.Split(inst[2], ",")
					sID, _ := strconv.Atoi(inst[4])

					if len(joinCommittee) > 0 && sID != chainID {
						panic(fmt.Sprint(sID, " diff ", chainID))
					}

					for i, _ := range joinCommittee {
						if joinCommittee[i] == "" {
							joinCommittee = []string{}
							continue
						}
						err := eventDB.createJoinCommitteeEvent(JoinCommitteeEvent{
							uint64(shardBlk.GetProduceTime()),
							shardBlk.GetCurrentEpoch(),
							joinCommittee[i],
							chainID,
							blkHeight,
							blkHash.String(),
						})
						if err != nil {
							panic(err)
						}
					}
					for i, _ := range leaveCommittee {
						if leaveCommittee[i] == "" {
							leaveCommittee = []string{}
							continue
						}
						err := eventDB.createLeaveCommitteeEvent(LeaveCommitteeEvent{
							uint64(shardBlk.GetProduceTime()),
							shardBlk.GetCurrentEpoch(),
							leaveCommittee[i],
							chainID,
							blkHeight,
							blkHash.String(),
						})
						if err != nil {
							panic(err)
						}
					}
				}
			}

			shardLastEpoch, ok := lastEpoch.Load(chainID)
			////fmt.Println(chainID, shardbBlk.GetHeight(), shardbBlk.Hash().String(), shardbBlk.GetCurrentEpoch(),shardLastEpoch, bc.ShardChain[chainID].GetFinalView().GetHeight(),bc.ShardChain[chainID].GetBestView().GetHeight())
			if (ok && shardBlk.GetCurrentEpoch() != shardLastEpoch) || (!ok && shardBlk.GetCurrentEpoch() == 1) {
				//	//store event
				err = eventDB.createStartEpoch(StartEpoch{
					uint64(shardBlk.GetProduceTime()),
					shardBlk.GetHeight(),
					shardBlk.GetCurrentEpoch(),
					shardBlk.Hash().String(),
					chainID,
				})
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}

				//store committee
				if shardBlk.Header.BeaconHeight < config.Param().ConsensusParam.StakingFlowV2Height {
					committeeState, err := remoteRPC.GetCommitteeStateOnShard(shardBlk.GetShardID(), blkHash.String())
					if err != nil {
						fmt.Println(shardBlk.GetShardID(), blkHash.String())
						panic("Cannot get committee state on shard")
					}

					err = committeeDB.saveCommittee(chainID, blkHeight, shardBlk.Header.Epoch, committeeState.Committee, uint64(shardBlk.GetProduceTime()), shardBlk.Hash().String())
					if err != nil {
						panic(err)
					}

				}
			}
			lastEpoch.Store(chainID, shardBlk.GetCurrentEpoch())
		})
	}
	select {}
}
