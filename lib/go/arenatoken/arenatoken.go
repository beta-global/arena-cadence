package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/flow-go-sdk"
)

// ArenaToken is the API handle for contructing transactions/scripts
// to interact with the token contract
type ArenaToken struct {
	contracts map[string]flow.Address
}

func New(contractAddr, fungibleTokenAddr flow.Address) *ArenaToken {
	contracts := make(map[string]flow.Address)
	contracts["ArenaToken"] = contractAddr
	contracts["FungibleToken"] = fungibleTokenAddr

	return &ArenaToken{contracts: contracts}
}

// Contract returns the source for deploying the ArenaToken fungible token contract
func Contract(fungibleTokenAddr flow.Address) string {
	contracts := map[string]flow.Address{"FungibleToken": fungibleTokenAddr}
	return arenacadence.Render(contractTemplate, nil, contracts)
}
