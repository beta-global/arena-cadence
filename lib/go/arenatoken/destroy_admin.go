package arenatoken

import (
	_ "embed"

	"github.com/onflow/flow-go-sdk"
)

// DestroyAdministrator returns an unsigned transaction for destroying the singular
// Admin resource. This will prevent any future minters from being created.
func (r *ArenaToken) DestroyAdministrator() *flow.Transaction {
	tx := render(destroyAdministratorTemplate, nil, r.contracts)

	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(40)
}
