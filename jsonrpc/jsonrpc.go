package jsonrpc

import "encoding/json"

func getIdStr(rawid *json.RawMessage) string {
	if rawid == nil {
		return ""
	}
	return string(*rawid)
}

func wrapAny (any interface{}) (*json.RawMessage, error) {
	if  res, ok := any.(*json.RawMessage); ok {
		return res, nil
	}

	var bytesid []byte
	var err error

	if bytesid, err = json.Marshal(any); err != nil {
		return nil, err
	}

	var res = json.RawMessage(bytesid)
	return &res, nil
}

func WrapNotification (method string, msg interface{}) ([]byte, error) {
	var err error
	var rawmsg *json.RawMessage

	if rawmsg, err = wrapAny(msg); err != nil {
		return nil, err
	}

	var nt = MessageHeader {
		Jsonrpc: "2.0",
		Method:  method,
		Params:  rawmsg,
	}

	var bres []byte
	if bres, err = json.Marshal(&nt); err != nil {
		return nil, err
	}

	return bres, nil
}

func WrapResponse (id interface{}, msg interface{}) ([]byte, error) {
	var rawid, rawmsg *json.RawMessage
	var err error

	if rawid, err = wrapAny(id); err != nil {
		return nil, err
	}

	if rawmsg, err = wrapAny(msg); err != nil {
		return nil, err
	}

	var resp = ResponseHeader {
		Jsonrpc: "2.0",
		Id:      rawid,
		Result:  rawmsg,
	}

	var response []byte
	if response, err = json.Marshal(&resp); err != nil {
		return nil, err
	}

	return response, nil
}



/*
// WrapError returns []byte representation of the error response
// This function never fails
// id can be string or *json.RawMessage or nil (interpreted as "-1")
func WrapError (id interface{}, code int, message string) [] byte {

	if id == nil {
		id = "-1"
	}
	// We go here without using ErrorHeader and avoid json marshaling
	// to be sure that this function never fails itself
	var errFmt = `{"jsonrpc":"2.0", "id": "%v", "error": {"code": %v, "message": "%v"}}`
	var rpcError = fmt.Sprintf(errFmt, code, err.Error())

	log.Printf("WARNING: jsonrpc error: %v", rpcError)
	return  []byte(rpcError)
}
*/