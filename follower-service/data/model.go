package data

import (
	"encoding/json"
	"io"
)

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	FollowedByMe bool   `json:"followedByMe"`
}

type Users []*User

func (o *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *User) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
