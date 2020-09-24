package wallet

import "encoding/json"

type sendResult struct {
	TxID string `json:"txId"`
}

func (api* API) SendBEAM(to string, from string, amount uint64, fee uint64) (txid string, err error) {
	var params = JsonParams{
		"value": amount,
		"fee": fee,
		"address": to,
		"from": from,
		"comment": "user withdraw request",
	}

	var rawres *json.RawMessage
	if rawres, err = api.rpcPost("tx_send", params); err != nil {
		return
	}

	var res = sendResult{}
	if err = json.Unmarshal(*rawres, &res); err != nil {
		return
	}

	txid = res.TxID
	return
}
