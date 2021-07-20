package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/flow-go-sdk"
)

// DestroyAdministrator returns an unsigned transaction for destroying the singular
// Admin resource. This will prevent any future minters from being created.
func (r *ArenaToken) DestroyAdministrator() *flow.Transaction {
	tx := arenacadence.Render(destroyAdministratorTemplate, nil, r.contracts)

	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(40)
}
