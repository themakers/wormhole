package decparser

import "github.com/themakers/wormhole/decparser/types"

type Result struct {
	Implicit    []types.Type
	Definitions []*types.Definition
	Packages    []*types.Package
}
