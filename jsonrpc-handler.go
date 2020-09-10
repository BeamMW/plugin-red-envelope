package main

import (
	"encoding/json"
	"fmt"
	"github.com/olahol/melody"
)

func onClientMessage (session *melody.Session, msg []byte) (response []byte) {
	return jsonRpcProcess(msg,
		func(method string, params *json.RawMessage) (result interface{}, errCode RpcErrCode, err error) {

			switch method {
			case "login":
				result, err = onClientLogin(session, params)
			case "logout":
				result, err = onClientLogout(session, params)
			case "get-status":
				result, err = onGetStatus(session, params)
			default:
				err = fmt.Errorf("method '%v' not found", method)
				errCode = NoMethod
			}

			if err != nil {
				if _, ok := err.(*json.MarshalerError); ok {
					errCode = ParseError
				} else {
					errCode = InternalError
				}
			}
			return
		})
}
