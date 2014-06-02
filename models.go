package ngAuthApi

import (
	"encoding/json"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mrvdot/appengine/aeutils"

	"appengine"
	"appengine/datastore"
)

var (
	usersBySession = map[string]*User{}
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Session  string `json:"session"`
}

func getBySession(ctx appengine.Context, session string) *User {
	if user, ok := usersBySession[session]; ok {
		return user
	}
	iter := datastore.NewQuery("User").
		Filter("Session =", session).
		Limit(1).
		Run(ctx)

	u := &User{}
	_, err := iter.Next(u)
	if err != nil {
		ctx.Errorf("Error loading user: %v", err.Error())
		return nil
	}
	return u
}

func newUser(ctx appengine.Context, req *http.Request) *User {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	u := &User{}
	err := dec.Decode(u)
	if err != nil {
		ctx.Errorf("Error decoding new user: %v", err.Error())
		return nil
	}
	aeutils.Save(ctx, u)
	return u
}

func (user *User) BeforeSave(ctx appengine.Context) {
	if user.Session == "" {
		user.Session = uuid.NewRandom().String()
	}
}

func (user *User) AfterSave(ctx appengine.Context, key *datastore.Key) {
	usersBySession[user.Session] = user
}
