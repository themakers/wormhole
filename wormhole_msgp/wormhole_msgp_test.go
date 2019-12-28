package wormhole_msgp

import (
	"bytes"
	"github.com/themakers/wormhole/wormhole/wire_io"
	"testing"
)

func TestIO(t *testing.T) {
	perror := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	buf := bytes.NewBuffer([]byte{})

	perror(Handler.NewWriter(2, buf, func(w wire_io.ValueWriter) error {
		perror(w.WriteString("Hello"))
		perror(w.WriteInt(127))
		return nil
	}))

	data := buf.Bytes()
	t.Log(data)

	sz, r, err := Handler.NewReader(bytes.NewReader(data))
	perror(err)

	if sz != 2 {
		panic("size mismatch")
	}

	v, err := r()
	perror(err)

	t.Log(v.(string))

	v, err = r()
	perror(err)

	t.Log(v.(int))
}
