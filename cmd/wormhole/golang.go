package main

import (
	"bytes"
	"github.com/themakers/wormhole/parsex"
	"text/template"
)

func Render(pkg string, ifcs []*parsex.Interface) string {

	model := struct {
		Pkg  string
		Ifcs []*parsex.Interface
	}{
		Pkg:  pkg,
		Ifcs: ifcs,
	}

	t, err := template.New("golang").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer([]byte{})

	if err := t.Execute(buf, model); err != nil {
		panic(err)
	}

	return buf.String()
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
	"time"
	"github.com/themakers/wormhole/wormhole"
)

{{range $i, $ifc := .Ifcs}}
/****************************************************************
** {{$ifc.Name}} Client
********/

	var _ {{$ifc.Name}} = (*{{template "clientImplStructName" $ifc}})(nil)

	type {{template "clientImplStructName" $ifc}} struct {
		peer wormhole.RemotePeer
	}

	func {{template "clientImplConstructorName" $ifc}}(peer wormhole.RemotePeer) {{$ifc.Name}} {
		return &{{template "clientImplStructName" $ifc}}{peer: peer}
	}

	{{range $i, $fn := $ifc.Methods}}
		func (impl *{{template "clientImplStructName" $ifc}}) {{$fn.Name}}(ctx context.Context, arg {{$fn.Arg}}) (ret {{$fn.Ret}}, err error) {
			return ret, impl.peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("{{$ifc.Name}}", "{{$fn.Name}}", reflect.TypeOf(impl.{{$fn.Name}}), ctx, arg, &ret)
		}
	{{end}}


/****************************************************************
** {{$ifc.Name}} Client (KeepAlive)
********/

	var _ {{$ifc.Name}} = (*{{template "clientKAImplStructName" $ifc}})(nil)

	type {{template "clientKAImplStructName" $ifc}} struct {
		peer wormhole.LocalPeer
		id   string
		to   time.Duration
	}

	func {{template "clientKAImplConstructorName" $ifc}}(peer wormhole.LocalPeer, id string, to time.Duration) {{$ifc.Name}} {
		return &{{template "clientKAImplStructName" $ifc}}{peer: peer, id: id, to: to}
	}

	{{range $i, $fn := $ifc.Methods}}
		func (impl *{{template "clientKAImplStructName" $ifc}}) {{$fn.Name}}(ctx context.Context, arg {{$fn.Arg}}) (ret {{$fn.Ret}}, err error) {
			waitCtx, cancel := context.WithTimeout(ctx, impl.to)
			defer cancel()
			if peer := impl.peer.(wormhole.LocalPeerGenerated).WaitFor(waitCtx, impl.id); peer != nil {
				return ret, peer.(wormhole.RemotePeerGenerated).MakeRootOutgoingCall("{{$ifc.Name}}", "{{$fn.Name}}", reflect.TypeOf(impl.{{$fn.Name}}), ctx, arg, &ret)
			} else {
				return ret, wormhole.ErrTimeout
			}
		}
	{{end}}

/****************************************************************
** {{$ifc.Name}} Handler
********/

	func {{template "serverProxyFuncName" $ifc}}(peer wormhole.LocalPeer, constructor func(caller wormhole.RemotePeer) {{$ifc.Name}}) {
		peer.(wormhole.LocalPeerGenerated).RegisterInterface("{{$ifc.Name}}", func(caller wormhole.RemotePeer) {
			ifc := constructor(caller)
			val := reflect.ValueOf(ifc)
			
			{{range $i, $fn := $ifc.Methods}}
			caller.(wormhole.RemotePeerGenerated).RegisterRootRef("{{$ifc.Name}}", "{{$fn.Name}}", val.MethodByName("{{$fn.Name}}")) {{end}}
		})
	}

{{end}}

`