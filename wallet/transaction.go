package wallet

import (
	"encoding/json"
	"fmt"
)

const (
	DefaultFee uint64 = 100000
)

type TxStatus uint32

const (
	TxPending TxStatus = iota
	TxInProgress
	TxCancelled
	TxCompleted
	TxFailed
	TxRegistering
)

type Transaction struct {
	TxId     string   `json:"txId"`
	Status   TxStatus `json:"status"`
	Value    uint64   `json:"value"`
	Receiver string   `json:"receiver"`
	Height   uint64   `json:"height"`
	Income   bool     `json:"income"`
	Fee      uint64   `json:"fee"`
}

func (api *API) GetTransactions() (txs []Transaction, err error) {
	var res *json.RawMessage
	if res, err = api.rpcPost("tx_list", nil); err != nil {
		return
	}

	err = json.Unmarshal(*res, &txs)
	return txs, nil
}

func (api *API) DeleteTransaction(txid string) error {
	var params = JsonParams{
		"txId":  txid,
	}

	var rawres *json.RawMessage
	var err error

	if rawres, err = api.rpcPost("tx_delete", params); err != nil {
		return err
	}

	var boolRes bool
	if err = json.Unmarshal(*rawres, &boolRes); err != nil {
		return err
	}

	if !boolRes {
		return fmt.Errorf("failed to remove transaction %s", txid)
	}

	return nil
}
