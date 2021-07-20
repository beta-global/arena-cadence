package arenatoken

import (
	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
)

// Mint returns an unsigned transaction for minting new tokens. Only an account holding
// the singular Admin resource can execute this transaction
func (r *ArenaToken) MintTokens(recipient flow.Address, amount cadence.UFix64) *flow.Transaction {
	tx := render(mintArenaTemplate, nil, r.contracts)

	// convert args to cadence compatible forms
	var buf [cadence.AddressLength]byte
	copy(buf[:], recipient.Bytes())

	return flow.NewTransaction().
		AddRawArgument(jsoncdc.MustEncode(cadence.NewAddress(buf))).
		AddRawArgument(jsoncdc.MustEncode(amount)).
		SetScript([]byte(tx)).
		SetGasLimit(60)
}
