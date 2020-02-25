package types

import (
	"fmt"
	"strings"
)

func stringify(t Type) string {
	var (
		prev = map[*Definition]bool{}
		do   func(Type) string
	)

	do = func(t Type) string {
		switch n := t.(type) {
		case untyped:
			return "<untyped>"

		case *Array:
			return fmt.Sprintf(
				"[%d]%s",
				n.Len,
				do(n.Type),
			)

		case Builtin:
			return builtin2String(n)

		case *Chan:
			return fmt.Sprintf(
				"<chan>%s",
				do(n.Type),
			)

		case *Definition:
			if n.Std {
				return fmt.Sprintf(
					"<STD_DEF>%s.%s",
					do(n.Package),
					n.Name,
				)
			}

			if prev[n] {
				return fmt.Sprintf(
					"%s.%s",
					do(n.Package),
					n.Name,
				)
			}

			prev[n] = true
			return fmt.Sprintf(
				"%s.%s :: %s",
				do(n.Package),
				n.Name,
				do(n.Declaration),
			)

		case *Function:
			var (
				args    = make([]string, len(n.Args))
				results = make([]string, len(n.Results))
			)
			for i, arg := range n.Args {
				args[i] = do(arg.Type)
			}
			for i, result := range n.Results {
				results[i] = do(result.Type)
			}

			return fmt.Sprintf(
				"func(%s)(%s)",
				strings.Join(args, ","),
				strings.Join(results, ","),
			)

		case *Interface:
			var methods = make([]string, len(n.Methods))
			for i, method := range n.Methods {
				methods[i] = do(method)
			}
			return fmt.Sprintf(
				"<interface>{\n%s\n}",
				shift(strings.Join(methods, "\n")),
			)

		case *Map:
			return fmt.Sprintf(
				"<map>[%s]%s",
				do(n.Key),
				do(n.Value),
			)

		case *Method:
			return fmt.Sprintf(
				"<method> %s :: %s",
				n.Name,
				do(n.Signature),
			)

		case *Package:
			return fmt.Sprintf(
				"<pkg>\"%s\"",
				n.Info.PkgPath,
			)

		case *Pointer:
			return fmt.Sprintf(
				"<Pointer>%s",
				do(n.Type),
			)

		case *Slice:
			return fmt.Sprintf(
				"[]%s",
				do(n.Type),
			)

		case *Struct:
			fields := make([]string, len(n.Fields))
			for i, field := range n.Fields {
				if field.Embedded {
					fields[i] = fmt.Sprintf(
						"%s `%s`",
						do(field.Type),
						field.Tag,
					)
				} else {
					fields[i] = fmt.Sprintf(
						"%s %s `%s`",
						field.Name,
						do(field.Type),
						field.Tag,
					)
				}
			}

			return fmt.Sprintf(
				"<struct>{\n%s\n}",
				shift(strings.Join(fields, "\n")),
			)

		default:
			panic("Unknown type")
		}
	}

	return do(t)
}

func shift(s string) string {
	res := strings.Split(s, "\n")
	for i, s := range res {
		res[i] = "  " + s
	}
	return strings.Join(res, "\n")
}

func builtin2String(b Builtin) string {
	switch b {
	case Int:
		return "<builtin>int"
	case Int32:
		return "<builtin>int32"
	case Int64:
		return "<builtin>int64"
	case Uint:
		return "<builtin>uint"
	case Uint32:
		return "<builtin>uint32"
	case Uint64:
		return "<builtin>uint64"
	case Byte:
		return "<builtin>byte"
	case String:
		return "<builtin>string"
	case Rune:
		return "<builtin>rune"
	case Bool:
		return "<builtin>bool"
	case Error:
		return "<builtin>error"
	default:
		panic("Invalid builtin type")
	}
}
