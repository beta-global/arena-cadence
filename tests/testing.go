package tests

import (
	"context"
	"testing"

	"github.com/arena/arena-cadence/lib/go/arenatoken"
	"github.com/arena/arena-cadence/tests/emulator"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

func DeployContract(t *testing.T, em *emulator.Emulator, owner flow.Address, name string, source string) {
	if _, err := em.DeployContract(owner, name, source); err != nil {
		t.Fatalf("Deploying contract: %v", err)
	}
}

func AddAccount(t *testing.T, em *emulator.Emulator) flow.Address {
	newAcct, err := em.AddAccount()
	if err != nil {
		t.Fatalf("Adding account: %v", err)
	}
	return newAcct
}

func arenaBalance(t *testing.T, em *emulator.Emulator, target flow.Address) cadence.UFix64 {
	t.Helper()

	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])
	balanceScript, args := txRenderer.Balance(target)
	val, err := em.Client.ExecuteScriptAtLatestBlock(context.Background(), balanceScript, args)
	if err != nil {
		t.Fatalf("Reading balance: %v", err)
	}

	return val.(cadence.UFix64)
}

// Do i make a simple struct to wrap emulator for a cleaner API? probably since I'm already doing this to simplify testing
