package main

import "github.com/themakers/wormhole/defparser"

func Render(pkg *defparser.Result) string {

	// model := struct {
	// 	Pkg  string
	// 	Ifcs []*parsex.Interface
	// }{
	// 	Pkg:  pkg,
	// 	Ifcs: ifcs,
	// }

	// t, err := template.New("golang").Parse(tmpl)
	// if err != nil {
	// 	panic(err)
	// }

	// buf := bytes.NewBuffer([]byte{})

	// if err := t.Execute(buf, model); err != nil {
	// 	panic(err)
	// }

	// return buf.String()

	return ""
}

const tmpl = `
{{define "clientImplStructName"}} wormhole{{.Name}}ClientImpl {{end}}
{{define "clientImplConstructorName"}} Acquire{{.Name}} {{end}}
{{define "clientKAImplStructName"}} wormhole{{.Name}}KeepAliveClientImpl {{end}}
{{define "clientKAImplConstructorName"}} AcquireKeepAlive{{.Name}} {{end}}

{{define "serverProxyFuncName"}} Register{{.Name}}Handler {{end}}

package {{.Pkg}}

import (
	"context"
	"fmt"

	__wormhole "github.com/themakers/wormhole/wormhole"
	wormhole "github.com/themakers/wormhole/wormhole"
	__wire_io "github.com/themakers/wormhole/wormhole/wire_io"
	__reflect_hack "github.com/themakers/wormhole/wormhole/reflect_hack"
)

{{range $i, $ifc := .Ifcs}}
/****************************************************************
** {{$ifc.Name}} Client
********/

	var _ {{$ifc.Name}} = (*{{template "clientImplStructName" $ifc}})(nil)

	type {{template "clientImplStructName" $ifc}} struct {
		peer __wormhole.RemotePeer
	}

	func {{template "clientImplConstructorName" $ifc}}(peer __wormhole.RemotePeer) {{$ifc.Name}} {
		return &{{template "clientImplStructName" $ifc}}{peer: peer}
	}

	{{range $i, $fn := $ifc.Methods}}
		func (__impl *{{template "clientImplStructName" $ifc}}) {{$fn.String}} {
			__peer := __impl.peer.(__wormhole.RemotePeerGenerated)

			__doneCtx, __done := context.WithCancel(ctx)
			defer __done()

			__peer.Call("{{$ifc.Name}}.{{$fn.Name}}", ctx, {{len $fn.Ins}}, func(__rr __wormhole.RegisterUnnamedRefFunc, __w __wire_io.ValueWriter) {
				{{range $fn.Ins}}
					__reflect_hack.WriteAny(__peer, __rr, __w, {{.}}){{end}}

			}, func(ctx context.Context, __ar __wire_io.ArrayReader) {
				defer __done()
				__sz, __r, __err := __ar()
				if __err != nil {
					panic(__err)
				}
				if __sz != {{len $fn.Outs}} {
					panic(fmt.Sprintf("return values count mismatch: %d != %d", {{len $fn.Outs}}, __sz))
				}

				{{range $fn.Outs}}
					__reflect_hack.ReadAny(__peer, __r, &{{.}}){{end}}

			})

			<-__doneCtx.Done()

			return
		}
	{{end}}

/****************************************************************
** {{$ifc.Name}} Handler
********/

	func {{template "serverProxyFuncName" $ifc}}(localPeer wormhole.LocalPeer, constructor func(wormhole.RemotePeer) {{$ifc.Name}}) {
		localPeer.(__wormhole.LocalPeerGenerated).RegisterInterface("{{$ifc.Name}}", func(peer __wormhole.RemotePeer) {
			__ifc := constructor(peer)
			__peer := peer.(__wormhole.RemotePeerGenerated)
			{{range $i, $fn := $ifc.Methods}}
			__peer.RegisterServiceRef("{{$ifc.Name}}.{{$fn.Name}}", func(ctx context.Context, __ar __wire_io.ArrayReader, __wf func(int, func(__wormhole.RegisterUnnamedRefFunc, __wire_io.ValueWriter))) {
				var ( {{range $arg := $fn.Args}}
					{{$arg.String}}{{end}}
				)

				__sz, __r, __err := __ar()
				if __err != nil {
					panic(__err)
				}
				if __sz != {{len $fn.Ins}} {
					panic(fmt.Sprintf("arguments count mismatch: %d != %d", {{len $fn.Ins}}, __sz))
				}

				{{range $fn.Ins}}
				__reflect_hack.ReadAny(__peer, __r, &{{.}}){{end}}

				{{$fn.RetsList}} := __ifc.{{$fn.Name}}(ctx, {{$fn.ArgsList}})

				__wf({{len $fn.Outs}}, func(__rr __wormhole.RegisterUnnamedRefFunc, __w __wire_io.ValueWriter) { {{range $fn.Outs}}
					__reflect_hack.WriteAny(__peer, __rr, __w, {{.}}){{end}}
				})
			}) {{end}}
		})
	}

{{end}}

`
