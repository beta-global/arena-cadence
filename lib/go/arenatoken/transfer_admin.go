package arenatoken

import (
	_ "embed"

	"github.com/onflow/flow-go-sdk"
)

// TransferAdministrator returns an unsigned transaction for handing over control of the
// singular Admin resource controlling the token contract
func (r *ArenaToken) TransferAdministrator(currentAdmin, newAdmin flow.Address) *flow.Transaction {
	tx := render(transferAdministratorTemplate, nil, r.contracts)

	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(40)
}
