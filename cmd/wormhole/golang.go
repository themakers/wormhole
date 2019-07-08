package wormhole

import (
	"bytes"
	"text/template"
)

func Render(pkg string, ifcs []Interface) string {

	model := struct {
		Pkg  string
		Ifcs []Interface
	}{
		Pkg:  pkg,
		Ifcs: ifcs,
	}

	t, err := template.New("golang").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer([]byte{})

	t.Execute(buf, model)

	return buf.String()
}

const tmpl = `
{{define "clientImplStructName"}} impl_client_{{.Name}} {{end}}
{{define "clientImplConstructorName"}} Acquire{{.Name}} {{end}}

{{define "fnArgs"}}{{range $i, $arg := .Args}} {{$arg.Name}} {{$arg.Type}}, {{end}}{{end}}
{{define "fnRets"}}{{range $i, $ret := .Rets}} {{$ret.Name}} {{$ret.Type}}, {{end}}{{end}}

{{define "fnArgsToCall"}}{{range $i, $arg := .Args}} {{$arg.Name}}, {{end}}{{end}}
{{define "fnRetsToCall"}}{{range $i, $ret := .Rets}} &{{$ret.Name}}, {{end}}{{end}}


{{define "serverProxyFuncName"}} Register{{.Name}}Handler {{end}}

package {{.Pkg}}

import (
	"github.com/themakers/nowire/nowire"
)

{{range $i, $ifc := .Ifcs}}
/****************************************************************
** {{$ifc.Name}} Client
********/

	var _ {{$ifc.Name}} = (*{{template "clientImplStructName" $ifc}})(nil)

	type {{template "clientImplStructName" $ifc}} struct {
		peer nowire.RemotePeer
	}

	func {{template "clientImplConstructorName" $ifc}}(peer nowire.RemotePeer) {{$ifc.Name}} {
		return &{{template "clientImplStructName" $ifc}}{peer: peer}
	}

	{{range $i, $fn := $ifc.Methods}}
		func (impl *{{template "clientImplStructName" $ifc}}) {{$fn.Name}}({{template "fnArgs" $fn}}) ({{template "fnRets" $fn}}) {
			mtype, _ := reflect.TypeOf(impl).MethodByName("{{$fn.Name}}")
			impl.peer.(nowire.RemotePeerGenerated).MakeOutgoingCall("{{$ifc.Name}}", "{{$fn.Name}}", mtype.Type, []interface{}{ {{template "fnArgsToCall" $fn}} }, []interface{}{ {{template "fnRetsToCall" $fn}} })
			return
		}
	{{end}}

/****************************************************************
** {{$ifc.Name}} Handler
********/

	func {{template "serverProxyFuncName" $ifc}}(peer nowire.LocalPeer, constructor func(caller nowire.RemotePeer) {{$ifc.Name}}) {
		peer.(nowire.LocalPeerGenerated).RegisterInterface("{{$ifc.Name}}", func(caller nowire.RemotePeer) {
			ifc := constructor(caller)
			val := reflect.ValueOf(ifc)
			
			{{range $i, $fn := $ifc.Methods}}
			caller.(nowire.RemotePeerGenerated).RegisterMethod("{{$ifc.Name}}", "{{$fn.Name}}", val.MethodByName("{{$fn.Name}}")) {{end}}
		})
	}

{{end}}

`
