package wallet

import (
	"bytes"
	"encoding/json"
	"github.com/BeamMW/red-envelope/jsonrpc"
	"io/ioutil"
	"net/http"
)

type JsonParams map[string]interface{}

func (api *API) rpcPost(method string, params interface{}) (res *json.RawMessage, err error) {
	var pbytes []byte
	if pbytes, err = json.Marshal(params); err != nil {
		return
	}

	rawId := json.RawMessage(`"http-dummy"`)
	rawParams := json.RawMessage(pbytes)

	request := jsonrpc.RPCRequest{
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
	if resp, err = http.Post(api.Address, "application/json", rbuffer); err != nil {
		return
	}

	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	var rpcr jsonrpc.RPCResponse
	if err = json.Unmarshal(body, &rpcr); err != nil {
		return
	}

	res = rpcr.Result
	return
}
