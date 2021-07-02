package arenatoken

import (
	"fmt"

	arenacadence "github.com/arena/arena-cadence"
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

func (r *Renderer) MintTokens(recipient flow.Address, amount uint64) (*flow.Transaction, error) {
	tx := arenacadence.Render(mintArenaTemplate, nil, r.contracts)

	// convert args to cadence compatible forms
	var buf [cadence.AddressLength]byte
	copy(buf[:], recipient.Bytes())

	amtFix, err := cadence.NewUFix64FromParts(int(amount), 0)
	if err != nil {
		return nil, fmt.Errorf("Provided amount is not a valid UFix64 value: %v", err)
	}

	fmt.Printf("Debug: %s\n", amtFix.String())

	return flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(cadence.NewAddress(buf))).
		AddRawArgument(jsoncdc.MustEncode(amtFix)).
		SetScript([]byte(tx)).
		SetGasLimit(100), nil
}
