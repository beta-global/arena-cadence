// Burn returns an unsigned transaction for burning the provided amount. The calling
package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

// account must be in control of a Burner resource.
func (r *ArenaToken) Burn(amount cadence.UFix64) *flow.Transaction {
	tx := arenacadence.Render(burnTemplate, nil, r.contracts)

	return flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(amount)).
		SetScript([]byte(tx)).
		SetGasLimit(40)
}
