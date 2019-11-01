package json_format

import (
	"encoding/json"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/internal/proto"
	"log"
	"reflect"
)

////////////////////////////////////////////////////////////////
//// wrapper
////

type messageType string

const (
	typeCall   messageType = "call"
	typeResult messageType = "result"
)

type wrapper struct {
	Type    messageType     `json:"Type"`
	Payload json.RawMessage `json:"Payload"`
}

////////////////////////////////////////////////////////////////
////
////

var _ wormhole.WireFormatHandler = new(handler)

type handler struct{}

func New() wormhole.WireFormatHandler {
	return &handler{}
}

func (handler) Marshal(v interface{}) ([]byte, error) {
	wr := &wrapper{}

	switch pv := v.(type) {
	case *proto.CallMsg:
		wr.Type = typeCall
		wire := &CallMsg{
			ID:   pv.ID,
			Ref:  pv.Ref,
			Meta: pv.Meta,
			Vars: make([][]interface{}, len(pv.Vars)),
		}
		for i, v := range pv.Vars {
			if v[1].CanInterface() {
				wire.Vars[i] = []interface{}{v[0].Interface(), v[1].Interface()}
			}
		}
		v = wire
	case *proto.ResultMsg:
		wr.Type = typeResult
		wire := &ResultMsg{
			Call: pv.Call,
			Meta: pv.Meta,
			Result: Result{
				Error: pv.Result.Error,
				Vals:  make([][]interface{}, len(pv.Result.Vals)),
			},
		}
		for i, v := range pv.Result.Vals {
			if v[1].CanInterface() {
				wire.Result.Vals[i] = []interface{}{v[0].Interface(), v[1].Interface()}
			}
		}
		v = wire
	default:
		panic("shit happened")
	}

	if data, err := json.Marshal(v); err != nil {
		return nil, err
	} else {
		wr.Payload = data
	}

	if data, err := json.Marshal(wr); err != nil {
		return nil, err
	} else {
		log.Println("Marshal", string(data))
		return data, nil
	}
}

func (handler) Unmarshal(m []byte) (interface{}, error) {

	log.Println("Unmarshal", string(m))

	var (
		wr wrapper
	)
	if err := json.Unmarshal(m, &wr); err != nil {
		return nil, err
	}

	switch wr.Type {
	case typeCall:
		var v CallMsg
		if err := json.Unmarshal(wr.Payload, &v); err != nil {
			return nil, err
		} else {
			pv := &proto.CallMsg{
				ID:   v.ID,
				Ref:  v.Ref,
				Meta: v.Meta,
			}
			for _, v := range v.Vars {
				pv.Vars = append(pv.Vars, []reflect.Value{reflect.ValueOf(v[0]), reflect.ValueOf(v[1])})
			}
			return pv, nil
		}
	case typeResult:
		var v ResultMsg
		if err := json.Unmarshal(wr.Payload, &v); err != nil {
			return nil, err
		} else {
			pv := &proto.ResultMsg{
				Call: v.Call,
				Meta: v.Meta,
				Result: proto.Result{
					Error: v.Result.Error,
				},
			}
			for _, v := range v.Result.Vals {
				pv.Result.Vals = append(pv.Result.Vals, []reflect.Value{reflect.ValueOf(v[0]), reflect.ValueOf(v[1])})
			}
			return pv, nil
		}
	default:
		panic("shit happened")
	}

	return nil, nil
}

////////////////////////////////////////////////////////////////
////
////

type CallMsg struct {
	ID string

	Ref string

	Meta map[string]string

	Vars [][]interface{}
}

type ResultMsg struct {
	Call string

	Meta map[string]string

	Result Result
}

type Result struct {
	Vals  [][]interface{}
	Error string
}
