package arenatoken

import (
	_ "embed"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

// Balance returns a script for fetching the ArenaToken balance of the provided account
func (r *ArenaToken) Balance(target flow.Address) ([]byte, []cadence.Value) {

	var arg cadence.Address
	copy(arg[:], target.Bytes())

	script := arenacadence.Render(balanceTemplate, nil, r.contracts)

	return []byte(script), []cadence.Value{arg}
}
