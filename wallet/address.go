package wallet

import "encoding/json"

type Expiration string

const (
	ExpNever Expiration = "never"
	Exp24h   Expiration = "24h"
)

type Address struct {
	Address   string `json:"address"`
	Comment   string `json:"comment"`
	Expired   bool   `json:"expired"`
	Duration  uint64 `json:"duration"`
}

type AddrList []Address

func (api* API) CreateAddress(comment string, expiration Expiration) (addr string, err error) {
	var params = JsonParams{
		"expiration": expiration,
		"comment": comment,
	}

	var res *json.RawMessage
	if res, err = api.rpcPost("create_address", params); err != nil {
		return
	}

	err = json.Unmarshal(*res, &addr)
	return
}

func (api* API) OwnAddrList() (list AddrList, err error) {
	var params = JsonParams{
		"own": true,
	}

	var res *json.RawMessage
	if res, err = api.rpcPost("addr_list", params); err != nil {
		return
	}

	list = make(AddrList, 0)
	err = json.Unmarshal(*res, &list)

	return
}
