package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type WalletAPI struct {
}

type JsonParams map[string]interface{}
func (api* WalletAPI) jsonRpcPost(method string, params interface {}) (res *json.RawMessage, err error) {
	var pbytes []byte
	if pbytes, err = json.Marshal(params); err != nil {
		return
	}

	rawId := json.RawMessage(`"http-dummy"`)
	rawParams := json.RawMessage(pbytes)

	request := RPCRequest{
		Jsonrpc: "2.0",
		Id:      &rawId,
		Method:  method,
		Params:  &rawParams,
	}

	var rbytes []byte
	if rbytes, err = json.Marshal(request); err != nil {
		return
	}

	var rbuffer = bytes.NewBuffer(rbytes)
	var resp *http.Response
	if resp, err = http.Post(config.WalletAPIAddress, "application/json", rbuffer); err != nil {
		return
	}

	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	var rpcr RPCResponse
	if err = json.Unmarshal(body, &rpcr); err != nil {
		return
	}

	res = rpcr.Result
	return
}

func (api* WalletAPI) CreateAddress() (addr string, err error) {
	var params = JsonParams {
		"expiration": "never",
	}

	var res *json.RawMessage
	if res, err = api.jsonRpcPost("create_address", params); err != nil {
		return
	}

	err = json.Unmarshal(*res, &addr)
	return
}

type WalletStatus struct {
	Available uint64 `json:"available"`
	Receiving uint64 `json:"receiving"`
}

func (api* WalletAPI) GetStatus() (status WalletStatus, err error) {
	var res *json.RawMessage
	if res, err = api.jsonRpcPost("wallet_status", nil); err != nil {
		return
	}

 	err = json.Unmarshal(*res, &status)
	return
}

type TxStatus uint32
const (
	Pending TxStatus = iota
	InProgress
	Cancelled
	Completed
	Failed
	Registering
)

type Transaction struct {
	Status   TxStatus `json:"status"`
	Value    uint64   `json:"value"`
	Receiver string   `json:"receiver"`
}

func (api* WalletAPI) GetTransactions() (txs []Transaction, err error) {
	var res *json.RawMessage
	if res, err = api.jsonRpcPost("tx_list", nil); err != nil {
		return
	}

	err = json.Unmarshal(*res, &txs)
	return txs, nil
}

var (
	wallet = &WalletAPI{}
)
