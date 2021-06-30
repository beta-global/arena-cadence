package tests

import (
	"fmt"
	"testing"

	"github.com/arena/arena-cadence/lib/go/arenatoken"
	"github.com/arena/arena-cadence/tests/emulator"
	"github.com/onflow/flow-go-sdk"
)

func TestContractEmbed(t *testing.T) {
	deploy := arenatoken.Contract(flow.HexToAddress(emulator.FungibleTokenAddr))
	fmt.Println(deploy)
}

func TestContractDeploy(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569")
	defer teardown()

	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	if err := em.DeployContract(em.ServiceAccount, "ArenaToken", contractSource); err != nil {
		t.Fatalf("failed to deploy contract: %v", err)
	}
}

func TestCreateAccount(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569")
	defer teardown()

	newAcct := AddAccount(t, em)
	if newAcct != flow.EmptyAddress {
		t.Fatalf("Expected non-empty address")
	}

}
