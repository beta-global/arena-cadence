package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/flow-go-sdk"
)

type Renderer struct {
	contracts map[string]flow.Address
}

func NewRenderer(contractAddr, fungibleTokenAddr flow.Address) *Renderer {
	contracts := make(map[string]flow.Address)
	contracts["ArenaToken"] = contractAddr
	contracts["FungibleToken"] = fungibleTokenAddr

	return &Renderer{contracts: contracts}
}

// Contract returns the source for deploying the ArenaToken fungible token contract
func Contract(fungibleTokenAddr flow.Address) string {
	contracts := map[string]flow.Address{"FungibleToken": fungibleTokenAddr}
	return arenacadence.Render(contractTemplate, nil, contracts)
}
