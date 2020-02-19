package decparser

import "github.com/themakers/wormhole/decparser/types"

type Result struct {
	Definitions    []*types.Definition
	STDDefinitions []*types.Definition

	Packages    []*types.Package
	STDPackages []*types.Package

	Methods  []*types.Method
	Implicit []types.Type
}
