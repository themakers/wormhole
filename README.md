# wormhole

Nowadays RPC systems can pass **some** first-class objects between client/server.
Some of them (e.g. gRPC) introduces concepts like _streams_ and _oneof_ to cover more use cases.
But these concepts bring more complexity to RPC because they are unnatural to programming languages.

Maybe RPC should look more like regular procedure call?

If we could pass more first-class members, especially "pass" `function references` between remote sides,
the need of such unnatural concepts would be eliminated.

Look at this code:
```go
// client
resp, err := greeter.Hello(ctx, api.GreeterHelloReq{
    Message: "please yield some messages",
    Yield: func(ctx context.Context, data string) (string, error) {
    	log.Println(data)
        return "something", nil
    },
}
log.Println(resp.Message)
```
```go
// server
func (gr *greeterServer) Hello(ctx context.Context, q api.GreeterHelloReq) (api.GreeterHelloResp, error) {
	log.Println("message:", q.Message)
	
	for _, msg := range []string{"msg1", "msg2", "msg3"} {
	    q.Yield(ctx, msg)
	}

	return api.GreeterHelloResp{Message: "done!"}, nil
}
```

Looks promising?

`TODO` More demos in the `demo` folder  
`TODO` More docs to come
 

This repository contains both the `wormhole` tool and `Go` support library

#### How to use it

```bash
# go get -u github.com/themakers/wormhole
# go install github.com/themakers/wormhole/cmd/wormhole
# cd <your api definition>
# wormhole go
```

 * `TODO` Write about companion JS library ([github.com/themakers/wormhole.js](https://github.com/themakers/wormhole.js))

#### How it works
 * `TODO` Explain how we have adopted Einstein-Rosen bridge to bring the idea to life (video)

#### Status
`pre-alpha`; API is in flux, protocol also; In fact working proof-of-concept that we are actively developing 

Starting from `beta` we are going follow `semver` and tag releases.
*You could click `Watch` button and select `Releases only` to be notified when the system reaches beta.*
Or, of course, you can contribute!

#### How to contribute
Any help is appreciated!
If you want to contribute, please file an issue first to discuss your proposal. 
 
#### Who using it
 * `TODO`

#### TODO
 - [ ] Make args/results handler completely recursive to support such constructs: `struct {Field map[string]func()}`
 - [ ] Complete demos
 - [ ] WASM/GopherJS readiness
 - [ ] Implement handshake phase?
 - [ ] Pass Go channels?

#### Why?
 * GRPC
   * Can't return message and error simultaneously
   * `oneof` concept have no reflection in most languages
   * `stream` concept also have no direct reflection in most languages
   * NIH

 
