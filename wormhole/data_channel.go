package wormhole

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type DataChannel interface {
	Context() context.Context
	ReadMessage() (interface{}, error)
	WriteMessage(interface{}) error
	Close() error
}

type JSONDataChannel struct {
	ctx context.Context
	log *zap.Logger
	dc  RawDataChannel
}

func NewJSONDataChannel(ctx context.Context, log *zap.Logger, dc RawDataChannel) DataChannel {
	return &JSONDataChannel{log: log, dc: dc}
}

func (c *JSONDataChannel) Context() context.Context {
	return c.ctx
}

func (c *JSONDataChannel) ReadMessage() (interface{}, error) {
	c.log.Debug("JDC waiting for message")
	data, err := c.dc.ReadMessage()
	if err != nil {
		c.log.DPanic("Eror reading encoded message", zap.Error(err))
		return nil, err
	}
	c.log.Debug("JDC got message", zap.String("message", string(data)))

	var msg struct {
		Type    string
		Payload json.RawMessage
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		c.log.DPanic("Eror unmarshaling message", zap.Error(err))
		return nil, err
	}

	switch msg.Type {
	case "call":
		var call Call
		if err := json.Unmarshal(msg.Payload, &call); err != nil {
			c.log.DPanic("Eror unmarshaling call message", zap.Error(err))
			return nil, err
		} else {
			return &call, nil
		}
	case "result":
		var result Result
		if err := json.Unmarshal(msg.Payload, &result); err != nil {
			c.log.DPanic("Eror unmarshaling result message", zap.Error(err))
			return nil, err
		} else {
			return &result, nil
		}
	default:
		err := fmt.Errorf("Unknown message type: %s", msg.Type)
		c.log.DPanic("Eror unmarshaling message", zap.Error(err))
		return nil, err
	}
}

func (c *JSONDataChannel) WriteMessage(msg interface{}) error {
	c.log.Debug("JDC writing message", zap.Any("message", msg))
	var mt string
	switch msg.(type) {
	case Call, *Call:
		mt = "call"
	case Result, *Result:
		mt = "result"
	default:
		err := fmt.Errorf("Unknown message type: %t", msg)
		c.log.DPanic("Eror unmarshaling message", zap.Error(err))
		return err
	}

	if data, err := json.MarshalIndent(struct {
		Type    string
		Payload interface{}
	}{
		Type:    mt,
		Payload: msg,
	}, "", " "); err != nil {
		c.log.DPanic("Eror marshaling result message", zap.Error(err))
		return err
	} else {
		if err := c.dc.WriteMessage(data); err != nil {
			c.log.DPanic("Eror writing encoded message", zap.Error(err))
			return err
		}
	}
	return nil
}

func (c *JSONDataChannel) Close() error {
	return c.dc.Close()
}
