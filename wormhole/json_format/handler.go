package json_format

import (
	"encoding/json"
	"github.com/themakers/wormhole/wormhole"
	"github.com/themakers/wormhole/wormhole/internal/proto"
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
			Vars: make([]interface{}, len(pv.Vars)),
		}
		for i, v := range pv.Vars {
			if v.CanInterface() {
				wire.Vars[i] = v.Interface()
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
				Vals:  make([]interface{}, len(pv.Result.Vals)),
			},
		}
		for i, v := range pv.Result.Vals {
			if v.CanInterface() {
				wire.Result.Vals[i] = v.Interface()
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
		return data, nil
	}
}

func (handler) Unmarshal(m []byte) (interface{}, error) {
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
				pv.Vars = append(pv.Vars, reflect.ValueOf(v))
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
				pv.Result.Vals = append(pv.Result.Vals, reflect.ValueOf(v))
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

	Vars []interface{}
}

type ResultMsg struct {
	Call string

	Meta map[string]string

	Result Result
}

type Result struct {
	Vals  []interface{}
	Error string
}

//
//type JSONDataChannel struct {
//	ctx context.Context
//	log *zap.Logger
//	dc  base.RawDataChannel
//}
//
//func NewJSONDataChannel(ctx context.Context, log *zap.Logger, dc base.RawDataChannel) wormhole.DataChannel {
//	return &JSONDataChannel{ctx: ctx, log: log, dc: dc}
//}
//
//func (c *JSONDataChannel) Context() context.Context {
//	return c.ctx
//}
//
//func (c *JSONDataChannel) ReadMessage() (interface{}, error) {
//	c.log.Debug("JDC waiting for message")
//	data, err := c.dc.ReadMessage()
//	if err != nil {
//		c.log.DPanic("Eror reading encoded message", zap.Error(err))
//		return nil, err
//	}
//	c.log.Debug("JDC got message", zap.String("message", string(data)))
//
//	var msg struct {
//		Type    string
//		Payload json.RawMessage
//	}
//	if err := json.Unmarshal(data, &msg); err != nil {
//		c.log.DPanic("Eror unmarshaling message", zap.Error(err))
//		return nil, err
//	}
//
//	switch msg.Type {
//	case "call":
//		var call proto.call
//		if err := json.Unmarshal(msg.Payload, &call); err != nil {
//			c.log.DPanic("Eror unmarshaling call message", zap.Error(err))
//			return nil, err
//		} else {
//			return &call, nil
//		}
//	case "result":
//		var result proto.Result
//		if err := json.Unmarshal(msg.Payload, &result); err != nil {
//			c.log.DPanic("Eror unmarshaling result message", zap.Error(err))
//			return nil, err
//		} else {
//			return &result, nil
//		}
//	default:
//		err := fmt.Errorf("Unknown message type: %s", msg.Type)
//		c.log.DPanic("Eror unmarshaling message", zap.Error(err))
//		return nil, err
//	}
//}
//
//func (c *JSONDataChannel) WriteMessage(msg interface{}) error {
//	c.log.Debug("JDC writing message", zap.Any("message", msg))
//	var mt string
//	switch msg.(type) {
//	case proto.call, *proto.call:
//		mt = "call"
//	case proto.Result, *proto.Result:
//		mt = "result"
//	default:
//		err := fmt.Errorf("Unknown message type: %t", msg)
//		c.log.DPanic("Eror unmarshaling message", zap.Error(err))
//		return err
//	}
//
//	if data, err := json.MarshalIndent(struct {
//		Type    string
//		Payload interface{}
//	}{
//		Type:    mt,
//		Payload: msg,
//	}, "", " "); err != nil {
//		c.log.DPanic("Eror marshaling result message", zap.Error(err))
//		return err
//	} else {
//		if err := c.dc.WriteMessage(data); err != nil {
//			c.log.DPanic("Eror writing encoded message", zap.Error(err))
//			return err
//		}
//	}
//	return nil
//}
//
//func (c *JSONDataChannel) Close() error {
//	return c.dc.Close()
//}
//
//func marshal(v interface{}) json.RawMessage {
//	if data, err := json.Marshal(v); err != nil {
//		panic(err)
//	} else {
//		return json.RawMessage(data)
//	}
//}
//
//func unmarshal(data json.RawMessage, vout interface{}) {
//	switch vout := vout.(type) {
//	case reflect.Value:
//		v := reflect.New(vout.Type())
//		if err := json.Unmarshal(data, v.Interface()); err != nil {
//			panic(err)
//		}
//		vout.Set(v.Elem())
//	default:
//		if err := json.Unmarshal(data, vout); err != nil {
//			panic(err)
//		}
//	}
//}
