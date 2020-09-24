package wallet

import "encoding/json"

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
	Status   TxStatus `json:"status"`
	Value    uint64   `json:"value"`
	Receiver string   `json:"receiver"`
	Height   uint64   `json:"height"`
	Income   bool     `json:"income"`
	Fee      uint64   `json:"fee"`
}

func (api* API) GetTransactions() (txs []Transaction, err error) {
	var res *json.RawMessage
	if res, err = api.rpcPost("tx_list", nil); err != nil {
		return
	}

	err = json.Unmarshal(*res, &txs)
	return txs, nil
}
