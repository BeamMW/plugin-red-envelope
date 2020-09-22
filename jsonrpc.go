package jsonrpc

import (
	"encoding/json"
	"fmt"
	"log"
)

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
	ParseError      RpcErrCode = -32700
	InvalidRequest  RpcErrCode = -32600
	NoMethod        RpcErrCode = -32601
	BadMethodParams RpcErrCode = -32602
	InternalError   RpcErrCode = -32603
)

func getIdStr(rawid *json.RawMessage) string {
	if rawid == nil {
		return ""
	}
	return string(*rawid)
}

//  -> method name, json params -> error | rpc result
type RpcHandler func(string, *json.RawMessage) (interface{}, RpcErrCode, error)

func jsonRpcProcess(msg []byte, handler RpcHandler) (response []byte) {
	var err error
	var errCode RpcErrCode

	var requestId *json.RawMessage
	var requestResult interface{}

	defer func () {
		if err == nil {
			if requestResult != nil {
				var resp = RPCResponse{
					Jsonrpc: "2.0",
					Id:      requestId,
				}

				var bres []byte
				if bres, err = json.Marshal(requestResult); err == nil {
					rawmsg := json.RawMessage(bres)
					resp.Result = &rawmsg
					response, err = json.Marshal(resp)
					// do not return, we want to fall down to error handling
				}
			}
		}

		if err != nil {
			var errFmt = `{"jsonrpc":"2.0", "id": "-1", "error": {"code": %v, "message": "%v"}}`
			var rpcError = fmt.Sprintf(errFmt, errCode, err.Error())
			log.Printf("jsonrpc error: %v", rpcError)
			response = []byte(rpcError)
		}
	} ()

	var header = RPCHeader{}
	if err := json.Unmarshal(msg, &header); err != nil {
		errCode = ParseError
		return
	}

	if header.Error != nil {
		log.Printf("jsonrpc, received error response for id [%v], result %v", getIdStr(header.Id), string(*header.Error))
		return
	}

	if header.Result != nil {
		if config.Debug {
			log.Printf("jsonrpc, received response for id [%v], result %v", getIdStr(header.Id), string(*header.Result))
		}
		return
	}

	//
	// Assume we're dealing with request (method call)
	//
	if requestId = header.Id; header.Id == nil {
		errCode = InvalidRequest
		err = fmt.Errorf("missing jsonrpc id")
		return
	}

	if header.Jsonrpc != "2.0" {
		errCode = InvalidRequest
		err = fmt.Errorf("bad jsonrpc version %v", header.Jsonrpc)
		return
	}

	if header.Params == nil {
		err = errors.New("bad jsonrpc, params are nil")
		errCode = InvalidRequest
		return
	}

	if len(header.Method) == 0 {
		err = errors.New("bad jsonrpc, empty method")
		errCode = InvalidRequest
		return
	}

	requestResult, errCode, err = handler(header.Method, header.Params)
	return
}
