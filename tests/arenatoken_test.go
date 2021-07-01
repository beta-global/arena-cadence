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
	if newAcct == flow.EmptyAddress {
		t.Fatalf("Expected non-empty address")
	}
}

func TestSetupAccount(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569")
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)

	// create a new account and run the setup_account transaction
	newAcct := AddAccount(t, em)

	txRenderer := arenatoken.NewRenderer(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])
	tx := txRenderer.SetupAccount()
	signers := emulator.TxSigners{
		Proposer:    newAcct,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{newAcct},
	}
	em.SignTx(signers, tx)

	result := em.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("setup_account tx execution: %v", result.Error)
	}

}

func TestMintArena(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569")
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.NewRenderer(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	t.Run("MintToAdministrator", func(t *testing.T) {

		tx, err := txRenderer.MintTokens(em.ServiceAccount, 100)
		if err != nil {
			t.Fatalf("Setting up mint transaction: %v", err)
		}

		signers := emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{em.ServiceAccount},
		}
		em.SignTx(signers, tx)
		result := em.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("mint_arena tx execution: %v", result.Error)
		}
		// TODO(dave): balance check
	})

	t.Run("MintToNonAdmin", func(t *testing.T) {

		// create a new account and perform account setup
		newAcct := AddAccount(t, em)
		tx := txRenderer.SetupAccount()
		signers := emulator.TxSigners{
			Proposer:    newAcct,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{newAcct},
		}
		em.SignTx(signers, tx)
		result := em.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("setup_account tx execution: %v", result.Error)
		}

		// admin mints tokens to the newly setup account
		tx, err := txRenderer.MintTokens(newAcct, 100)
		if err != nil {
			t.Fatalf("Setting up mint transaction: %v", err)
		}

		signers = emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{em.ServiceAccount},
		}
		em.SignTx(signers, tx)
		result = em.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("mint_arena tx execution: %v", result.Error)
		}
		// TODO(dave): balance check
	})

	t.Run("MintInvalidAmount", func(t *testing.T) {

		newAcct := AddAccount(t, em)

		// overflow UFix64 should fail
		_, err := txRenderer.MintTokens(newAcct, 999999999999999999)
		if err == nil {
			t.Fatalf("Expected input sanitation to fail but did not")
		}
	})

	t.Run("MintToUninitializedAccount", func(t *testing.T) {

		// create a new account without a vault
		newAcct := AddAccount(t, em)

		// admin mints tokens to the newly setup account
		// Tx should revert because new account does not have a vault
		tx, err := txRenderer.MintTokens(newAcct, 100)
		if err != nil {
			t.Fatalf("Setting up mint transaction: %v", err)
		}

		signers := emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{em.ServiceAccount},
		}
		em.SignTx(signers, tx)
		result := em.ExecuteTxWaitForSeal(tx)
		if result.Error == nil {
			t.Fatalf("Expected mint to revert but did not")
		}
		// TODO(dave): balance check
	})

}
