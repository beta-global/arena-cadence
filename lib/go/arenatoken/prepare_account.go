package arenatoken

import (
	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/flow-go-sdk"
)

// SetupAccount returns an unsigned transaction that prepares a user account
// for sending and receiving ArenaTokens.
func (r *ArenaToken) SetupAccount() *flow.Transaction {
	tx := arenacadence.Render(setupAccountTemplate, nil, r.contracts)
	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(100)
}
