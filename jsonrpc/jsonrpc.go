package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

func getIdStr(rawid *json.RawMessage) string {
	if rawid == nil {
		return ""
	}
	return string(*rawid)
}

func ProcessMessage(msg []byte, debug bool, handler RpcHandler) (response []byte) {
	var err error
	var errCode RpcErrCode

	var requestId *json.RawMessage
	var requestResult interface{}

	defer func() {
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
			log.Printf("WARNING: jsonrpc error: %v", rpcError)
			response = []byte(rpcError)
		}
	}()

	var header = RPCHeader{}
	if err := json.Unmarshal(msg, &header); err != nil {
		errCode = ErrParse
		return
	}

	if header.Error != nil {
		log.Printf("WARNING: jsonrpc, received error response for id [%v], result %v", getIdStr(header.Id), string(*header.Error))
		return
	}

	if header.Result != nil {
		if debug {
			log.Printf("WARNING: jsonrpc, received response for id [%v], result %v", getIdStr(header.Id), string(*header.Result))
		}
		return
	}

	//
	// Assume we're dealing with request (method call)
	//
	if requestId = header.Id; header.Id == nil {
		errCode = ErrInvalidReq
		err = fmt.Errorf("missing jsonrpc id")
		return
	}

	if header.Jsonrpc != "2.0" {
		errCode = ErrInvalidReq
		err = fmt.Errorf("bad jsonrpc version %v", header.Jsonrpc)
		return
	}

	if header.Params == nil {
		err = errors.New("bad jsonrpc, params are nil")
		errCode = ErrInvalidReq
		return
	}

	if len(header.Method) == 0 {
		err = errors.New("bad jsonrpc, empty method")
		errCode = ErrInvalidReq
		return
	}

	requestResult, errCode, err = handler(header.Method, header.Params)
	return
}

func WrapMessage(id string, msg interface{}) ([]byte, error) {
	var bytesid []byte
	var err error

	if bytesid, err = json.Marshal(id); err != nil {
		return nil, err
	}

	var rawid = json.RawMessage(bytesid)
	var resp = RPCResponse{
		Jsonrpc: "2.0",
		Id:      &rawid,
	}

	var bres []byte
	if bres, err = json.Marshal(msg); err != nil {
		return nil, err
	}

	rawmsg := json.RawMessage(bres)
	resp.Result = &rawmsg

	var response []byte
	if response, err = json.Marshal(resp); err != nil {
		return nil, err
	}

	return response, nil
}
