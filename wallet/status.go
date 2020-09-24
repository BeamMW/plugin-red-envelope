package wallet

import "encoding/json"

type Status struct {
	Available uint64 `json:"available"`
	Receiving uint64 `json:"receiving"`
}

func (api* API) GetStatus() (status Status, err error) {
	var res *json.RawMessage
	if res, err = api.rpcPost("wallet_status", nil); err != nil {
		return
	}

	err = json.Unmarshal(*res, &status)
	return
}
