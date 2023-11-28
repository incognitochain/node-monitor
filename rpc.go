package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"io/ioutil"
	"net/http"
)

type RemoteRPCClient struct {
	Endpoint string
}

type CommitteeStateOnShard struct {
	Root       string   `json:"root"`
	ShardID    int      `json:"shardID"`
	Committee  []string `json:"committee"`
	Substitute []string `json:"substitute"`
}

type CommitteeStateOnBeacon struct {
	Root             string           `json:"root"`
	Committee        map[int][]string `json:"committee"`
	Substitute       map[int][]string `json:"substitute"`
	NextCandidate    []string         `json:"nextCandidate"`
	CurrentCandidate []string         `json:"currentCandidate"`
}

type ErrMsg struct {
	Code       int
	Message    string
	StackTrace string
}

func (r *RemoteRPCClient) sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *RemoteRPCClient) GetCommitteeStateOnBeacon(height uint64, hash string) (res CommitteeStateOnBeacon, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcommitteestate",
		"params":  []interface{}{height, hash},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	resp := struct {
		Result CommitteeStateOnBeacon
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)
	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}
	if err != nil {
		return res, err
	}
	return resp.Result, nil
}

func (r *RemoteRPCClient) GetCommitteeStateOnShard(shardID int, hash string) (res CommitteeStateOnShard, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcommitteestatebyshard",
		"params":  []interface{}{shardID, hash},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	resp := struct {
		Result CommitteeStateOnShard
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)
	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}
	if err != nil {
		return res, err
	}
	return resp.Result, nil

}

func (r *RemoteRPCClient) GetBlocksFromHeight(shardID int, from uint64, num int) (res interface{}, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblocksfromheight",
		"params":  []interface{}{shardID, from, num},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	if shardID == -1 {
		resp := struct {
			Result []types.BeaconBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	} else {
		resp := struct {
			Result []types.ShardBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	}
}

func (r *RemoteRPCClient) GetTransactionByHash(hash string) (res jsonresult.TransactionDetail, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "gettransactionbyhash",
		"params":  []interface{}{hash},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, err
	}

	type Alias struct {
		jsonresult.TransactionDetail
		Proof       []byte
		ProofDetail interface{}
	}
	resp := struct {
		Result Alias
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("xxxx2", string(body))
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		fmt.Println("xxxx3", resp.Error)
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result.TransactionDetail, nil
}

func (r *RemoteRPCClient) GetBeaconViewByHash(hash string) (res jsonresult.GetBeaconBestState, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbeaconviewbyhash",
		"params":  []interface{}{hash},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	resp := struct {
		Result jsonresult.GetBeaconBestState
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}
	if err != nil {
		return res, err
	}
	return resp.Result, nil
}
