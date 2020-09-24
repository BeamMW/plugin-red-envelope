package main

import (
	"fmt"
	"github.com/chapati/melody"
)

const (
	UserIDKey = "UserID"
)

func setUserID(session *melody.Session, id string) {
	session.Set(UserIDKey, id)
}

func getUserID(session *melody.Session) (string, error) {
	rawid, ok := session.Get(UserIDKey)
	if !ok {
		return "", fmt.Errorf("cannot get user id from session %v", session)
	}
	return rawid.(string), nil
}
