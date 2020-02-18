package decparser

import "github.com/themakers/wormhole/decparser/types"

type Result struct {
	Definitions []*types.Definition
	Packages    []*types.Package
	Methods     []*types.Method
	Implicit    []types.Type
}
