package main

import "github.com/olahol/melody"

const (
	UserIDKey = "UserID"
)

func setUserID(session *melody.Session, id string) {
	session.Set(UserIDKey, id)
}

func getUserID(session *melody.Session) string {
	rawid, ok := session.Get(UserIDKey)
	if !ok {
		panic("Cannot get user id from session")
	}

	return rawid.(string)
}
