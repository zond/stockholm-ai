package models

import (
	"appengine/datastore"
)

const (
	TurnKind = "Turn"
)

type Turns []Turn

type Turn struct {
	Id *datastore.Key
}
