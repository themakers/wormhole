package nowire

import (
	"encoding/json"
	"reflect"
)

type Call struct {
	ID string

	Method string

	Vars []json.RawMessage

	results chan []json.RawMessage
}

type Result struct {
	Call string
	Vars []json.RawMessage
}

func marshal(v interface{}) json.RawMessage {
	if data, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return json.RawMessage(data)
	}
}

func unmarshal(data json.RawMessage, vout interface{}) {
	switch vout := vout.(type) {
	case reflect.Value:
		v := reflect.New(vout.Type())
		if err := json.Unmarshal(data, v.Interface()); err != nil {
			panic(err)
		}
		vout.Set(v.Elem())
	default:
		if err := json.Unmarshal(data, vout); err != nil {
			panic(err)
		}
	}
}
