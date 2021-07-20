package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

// Transfer returns an unsigned transaction for transfering tokens to the provided account
func (r *ArenaToken) Transfer(recipient flow.Address, amount cadence.UFix64) *flow.Transaction {

	tx := arenacadence.Render(transferTemplate, nil, r.contracts)

	var buf cadence.Address
	copy(buf[:], recipient.Bytes())

	return flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(cadence.NewAddress(buf))).
		AddRawArgument(jsoncdc.MustEncode(amount)).
		SetScript([]byte(tx)).
		SetGasLimit(40)
}
