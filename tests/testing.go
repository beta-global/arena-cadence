package tests

import (
	"testing"

	"github.com/arena/arena-cadence/tests/emulator"
	"github.com/onflow/flow-go-sdk"
)

func DeployContract(t *testing.T, em *emulator.Emulator, owner flow.Address, name string, source string) {
	if err := em.DeployContract(owner, name, source); err != nil {
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

// Do i make a simple struct to wrap emulator for a cleaner API? probably since I'm already doing this to simplify testing
