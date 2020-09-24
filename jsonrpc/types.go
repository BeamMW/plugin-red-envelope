package jsonrpc

import "encoding/json"

type RPCHeader struct {
	Jsonrpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id"`
	Result  *json.RawMessage `json:"result"`
	Params  *json.RawMessage `json:"params"`
	Error   *json.RawMessage `json:"error"`
	Method  string           `json:"method"`
}

type RPCRequest struct {
	Jsonrpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id"` // TODO: do we need pointer here?
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
}

type RPCResponse struct {
	Jsonrpc string           `json:"jsonrpc"`
	Id      *json.RawMessage `json:"id"`
	Result  *json.RawMessage `json:"result"`
}

type RpcErrCode int

const (
	ErrParse      RpcErrCode = -32700
	ErrInvalidReq RpcErrCode = -32600
	ErrNoMethod   RpcErrCode = -32601
	ErrBadParams  RpcErrCode = -32602
	ErrInternal   RpcErrCode = -32603
)

//  method name, json params -> error | rpc result
type RpcHandler func(string, *json.RawMessage) (interface{}, RpcErrCode, error)
