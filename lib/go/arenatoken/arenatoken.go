package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
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

func (r *Renderer) Balance(target flow.Address) ([]byte, []cadence.Value) {

	var arg cadence.Address
	copy(arg[:], target.Bytes())

	script := arenacadence.Render(balanceTemplate, nil, r.contracts)

	return []byte(script), []cadence.Value{arg}
}

func (r *Renderer) Transfer(recipient flow.Address, amount cadence.UFix64) *flow.Transaction {

	tx := arenacadence.Render(transferTemplate, nil, r.contracts)

	var buf cadence.Address
	copy(buf[:], recipient.Bytes())

	return flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(cadence.NewAddress(buf))).
		AddRawArgument(jsoncdc.MustEncode(amount)).
		SetScript([]byte(tx)).
		SetGasLimit(40)
}

func (r *Renderer) TransferAdministrator(currentAdmin, newAdmin flow.Address) *flow.Transaction {
	tx := arenacadence.Render(transferAdministratorTemplate, nil, r.contracts)

	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(40)
}

func (r *Renderer) DestroyAdministrator() *flow.Transaction {
	tx := arenacadence.Render(destroyAdministratorTemplate, nil, r.contracts)

	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(40)
}
