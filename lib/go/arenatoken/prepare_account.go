package arenatoken

import (
	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/flow-go-sdk"
)

func (r *Renderer) SetupAccount() *flow.Transaction {
	tx := arenacadence.Render(setupAccountTemplate, nil, r.contracts)
	return flow.NewTransaction().
		SetScript([]byte(tx)).
		SetGasLimit(100)
}
