package server

import (
	"fmt"
	"strconv"
)

// Message encapsulates data sent and received via the websocket.
type Message struct {
	Kind     string  `json:"kind,omitempty"`
	Mutation string  `json:"mutation,omitempty"`
	Data     msgData `json:"data"`
}

type msgData map[string]interface{}

func (d msgData) getString(key string) (string, error) {
	value, ok := d[key].(string)
	if ok {
		return value, nil
	}
	return value, fmt.Errorf("Coudln't find key '%s' in Data %v", key, d)
}

func (d msgData) getInt64(key string) (int64, error) {
	found, ok := d[key]
	if !ok {
		return 0, fmt.Errorf("Couldn't find key '%s' in Data %v", key, d)
	}

	value, ok := found.(string)
	if !ok {
		return 0, fmt.Errorf("Couldn't transform value of key %s: '%#v' into int", key, found)
	}

	return strconv.ParseInt(value, 10, 64)
}

func (d msgData) getInt(key string) (int, error) {
	found, ok := d[key]
	if !ok {
		return 0, fmt.Errorf("Couldn't find key '%s' in Data %v", key, d)
	}

	value, ok := found.(string)
	if !ok {
		return 0, fmt.Errorf("Couldn't transform value of key %s: '%#v' into int", key, found)
	}

	return strconv.Atoi(value)
}

func (d msgData) getID() (int64, error) {
	strID, err := d.getString("id")
	if err != nil {
		logger.Println(err)
		return 0, err
	}

	return strconv.ParseInt(strID, 10, 64)
}
