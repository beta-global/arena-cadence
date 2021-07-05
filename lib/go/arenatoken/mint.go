package arenatoken

import (
	"fmt"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

func (r *Renderer) MintTokens(recipient flow.Address, amount cadence.UFix64) *flow.Transaction {
	tx := arenacadence.Render(mintArenaTemplate, nil, r.contracts)

	// convert args to cadence compatible forms
	var buf [cadence.AddressLength]byte
	copy(buf[:], recipient.Bytes())

	fmt.Println(recipient)
	fmt.Println(amount)

	return flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(cadence.NewAddress(buf))).
		AddRawArgument(jsoncdc.MustEncode(amount)).
		SetScript([]byte(tx)).
		SetGasLimit(100)
}
