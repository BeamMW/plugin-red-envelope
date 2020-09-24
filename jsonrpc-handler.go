package main

import (
	"encoding/json"
	"fmt"
	"github.com/BeamMW/red-envelope/jsonrpc"
	"github.com/chapati/melody"
)

func onClientMessage (session *melody.Session, msg []byte) (response []byte) {
	return jsonrpc.ProcessMessage(msg, config.Debug,
		func(method string, params *json.RawMessage) (result interface{}, errCode jsonrpc.RpcErrCode, err error) {

			switch method {
			case "login":
				result, err = onClientLogin(session, params)

			case "take":
				result, err = onClientTake(session, params)

			default:
				err = fmt.Errorf("method '%v' not found", method)
				errCode = jsonrpc.ErrNoMethod
			}

			if err != nil {
				if _, ok := err.(*json.MarshalerError); ok {
					errCode = jsonrpc.ErrParse
				} else {
					errCode = jsonrpc.ErrInternal
				}
			}
			return
		})
}
