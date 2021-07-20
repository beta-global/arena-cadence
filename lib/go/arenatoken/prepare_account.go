package arenatoken

import (
	"github.com/onflow/flow-go-sdk"
)

// SetupAccount returns an unsigned transaction that prepares a user account
// for sending and receiving ArenaTokens.
func (r *ArenaToken) SetupAccount() *flow.Transaction {
	tx := render(setupAccountTemplate, nil, r.contracts)
	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(100)
}
