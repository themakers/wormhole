package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/themakers/wormhole/tests/api"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

func BenchmarkHTTPBasic(b *testing.B) {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	var port int

	{
		lis, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}

		port = lis.Addr().(*net.TCPAddr).Port

		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, q *http.Request) {
			if err := json.NewEncoder(w).Encode(api.GreeterHelloResp{
				Message: "+",
			}); err != nil {
				panic(err)
			}
		})

		s := &http.Server{
			Handler: mux,
		}

		go func() {
			if err := s.Serve(lis); err != nil {
				panic(err)
			}
		}()
	}

	type GreeterHelloReq struct {
		Message string
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reqData, err := json.Marshal(GreeterHelloReq{
			Message: "+",
		})
		if err != nil {
			b.Log(err)
			b.FailNow()
		}

		resp, err := http.Post(fmt.Sprintf("http://localhost:%d/", port), "application/json", bytes.NewReader(reqData))
		if err != nil {
			b.Log(err)
			b.FailNow()
		}

		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			b.Log(err)
			b.FailNow()
		}

		if err := resp.Body.Close(); err != nil {
			b.Log(err)
			b.FailNow()
		}

		var respObj api.GreeterHelloResp
		if err := json.Unmarshal(respData, &respObj); err != nil {
			b.Log(err)
			b.FailNow()
		}
	}
}
