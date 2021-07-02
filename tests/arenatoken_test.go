package tests

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/arena/arena-cadence/lib/go/arenatoken"
	"github.com/arena/arena-cadence/tests/emulator"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

var dockerLogsOnFail = flag.Bool("dockerLogs", false, "Print docker container logs on test failure")

func TestContractEmbed(t *testing.T) {
	deploy := arenatoken.Contract(flow.HexToAddress(emulator.FungibleTokenAddr))
	fmt.Println(deploy)
}

func TestContractDeploy(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	if _, err := em.DeployContract(em.ServiceAccount, "ArenaToken", contractSource); err != nil {
		t.Fatalf("failed to deploy contract: %v", err)
	}
}

func TestCreateAccount(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	newAcct := AddAccount(t, em)
	if newAcct == flow.EmptyAddress {
		t.Fatalf("Expected non-empty address")
	}
}

func TestSetupAccount(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
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

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
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

		// Validate new balance
		bal := arenaBalance(t, em, em.ServiceAccount)
		if bal.String() != "69520.00000000" {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", "69520.00000000", bal.String())
		}
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

		// Validate new balance
		bal := arenaBalance(t, em, newAcct)
		if bal.String() != "100.00000000" {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", "100.00000000", bal.String())
		}
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
	})

}

func TestBalance(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.NewRenderer(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	t.Run("BalanceInitializedAccount", func(t *testing.T) {

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

		balanceScript, args := txRenderer.Balance(em.ServiceAccount)
		val, err := em.Client.ExecuteScriptAtLatestBlock(context.Background(), balanceScript, args)
		if err != nil {
			t.Fatalf("Reading balance: %v", err)
		}

		if val.(cadence.UFix64).String() != "69520.00000000" {
			t.Fatalf("Expected balance: %v, got: %v", "69520.00000000", val.(cadence.UFix64).String())
		}
	})

	t.Run("BalanceUninitializedAccount", func(t *testing.T) {

		newAcct := AddAccount(t, em)

		balanceScript, args := txRenderer.Balance(newAcct)
		_, err := em.Client.ExecuteScriptAtLatestBlock(context.Background(), balanceScript, args)
		// Expect script to fail because account does not have a vault
		if err == nil {
			t.Fatalf("Expected balance check to fail but did not")
		}
	})

}

func TestTransfer(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.NewRenderer(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	t.Run("TransferInitializedAccount", func(t *testing.T) {

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

		// admin transfer tokens to the newly setup account
		amount, _ := cadence.NewUFix64("100.0")
		tx = txRenderer.Transfer(newAcct, amount)
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

		// Validate new balance
		bal := arenaBalance(t, em, newAcct)
		if bal != amount {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", amount, bal)
		}

		// Transfer some back
		amount, _ = cadence.NewUFix64("60.0")
		tx = txRenderer.Transfer(em.ServiceAccount, amount)
		signers = emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       newAcct,
			Authorizers: []flow.Address{newAcct},
		}
		em.SignTx(signers, tx)
		result = em.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("mint_arena tx execution: %v", result.Error)
		}

		// Validate new balances
		expect, _ := cadence.NewUFix64("40.0")
		bal = arenaBalance(t, em, newAcct)
		if bal != expect {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", expect, bal)
		}
	})

	t.Run("TransferUnitializedAccount", func(t *testing.T) {

		// create a new account and attempt to transfer tokens
		newAcct := AddAccount(t, em)
		amount, _ := cadence.NewUFix64("100.0")

		// admin transfer tokens to the newly setup account
		tx := txRenderer.Transfer(newAcct, amount)
		signers := emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{em.ServiceAccount},
		}
		em.SignTx(signers, tx)
		result := em.ExecuteTxWaitForSeal(tx)

		// Expect tx to revert because user hadn't performed setup
		if result.Error == nil {
			t.Fatalf("expected transfer to fail but did not")
		}
	})

	t.Run("TransferExceedBalance", func(t *testing.T) {

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

		// attempt to transfer more tokens than the service account owns
		amount, err := cadence.NewUFix64("99999999999.0")
		if err != nil {
			t.Fatalf("Invalid UFix64 amount: %v", err)
		}
		tx = txRenderer.Transfer(newAcct, amount)
		signers = emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{em.ServiceAccount},
		}
		em.SignTx(signers, tx)
		result = em.ExecuteTxWaitForSeal(tx)

		// Expect tx to revert because user didn't have enough tokens
		if result.Error == nil {
			t.Fatalf("expected transfer to fail but did not")
		}
	})
}
