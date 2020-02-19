package defparser

import "github.com/themakers/wormhole/defparser/types"

type Result struct {
	Root *types.Package

	Definitions    []*types.Definition
	STDDefinitions []*types.Definition

	Packages    []*types.Package
	STDPackages []*types.Package

	Methods []*types.Method
	Types   []types.Type
}
