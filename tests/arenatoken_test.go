package tests

import (
	"context"
	"flag"
	"strings"
	"testing"

	"github.com/arena/arena-cadence/lib/go/arenatoken"
	"github.com/arena/arena-cadence/tests/emulator"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

var dockerLogsOnFail = flag.Bool("dockerLogs", false, "Print docker container logs on test failure")

const initialBalance = "100000000000.00000000"

func TestContractDeploy(t *testing.T) {
	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	if _, err := em.DeployContract(em.ServiceAccount, "ArenaToken", contractSource); err != nil {
		t.Fatalf("failed to deploy contract: %v", err)
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

	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])
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
	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	t.Run("MintToAdministrator", func(t *testing.T) {

		amt, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.MintTokens(em.ServiceAccount, amt)
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
		if bal.String() != "100000000100.00000000" {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", "100000000100.00000000", bal.String())
		}

		// check expected events
		validateEvents(t, result, []string{
			"MinterCreated",
			"TokensMinted",
			"TokensDeposited",
		})

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
		amt, _ := cadence.NewUFix64("100.0")
		tx = txRenderer.MintTokens(newAcct, amt)
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
		if bal.String() != amt.String() {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", amt.String(), bal.String())
		}

		// check expected events
		validateEvents(t, result, []string{
			"MinterCreated",
			"TokensMinted",
			"TokensDeposited",
		})

	})

	t.Run("MintToUninitializedAccount", func(t *testing.T) {

		// create a new account without a vault
		newAcct := AddAccount(t, em)

		// admin mints tokens to the newly setup account
		// Tx should revert because new account does not have a vault
		amt, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.MintTokens(newAcct, amt)

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

func TestBurn(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	t.Run("Burn", func(t *testing.T) {

		oldBalance := arenaBalance(t, em, em.ServiceAccount)

		amt, _ := cadence.NewUFix64("10.0")
		tx := txRenderer.Burn(amt)
		signers := emulator.TxSigners{
			Proposer:    em.ServiceAccount,
			Payer:       em.ServiceAccount,
			Authorizers: []flow.Address{em.ServiceAccount},
		}
		em.SignTx(signers, tx)
		result := em.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("burn_arena tx execution: %v", result.Error)
		}

		// Validate new balance
		newBalance := arenaBalance(t, em, em.ServiceAccount)
		target := oldBalance - amt
		if newBalance != target {
			t.Fatalf("Incorrect balance after burning, expected: %s, got: %s", target, newBalance)
		}

		// check expected events
		validateEvents(t, result, []string{
			"TokensWithdrawn",
			"BurnerCreated",
			"TokensBurned",
		})

	})

	t.Run("NonAdminBurn", func(t *testing.T) {

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

		// Have user account attempt create burner. Should revert
		amt, _ := cadence.NewUFix64("10.0")
		tx = txRenderer.Burn(amt)
		signers = emulator.TxSigners{
			Proposer:    newAcct,
			Payer:       newAcct,
			Authorizers: []flow.Address{newAcct},
		}
		em.SignTx(signers, tx)
		result = em.ExecuteTxWaitForSeal(tx)
		if result.Error == nil {
			t.Fatalf("expected burn to revert but did not")
		}

	})
}

func TestBalance(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	t.Run("BalanceInitializedAccount", func(t *testing.T) {

		amt, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.MintTokens(em.ServiceAccount, amt)

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

		if val.(cadence.UFix64).String() != "100000000100.00000000" {
			t.Fatalf("Expected balance: %v, got: %v", "100000000100.00000000", val.(cadence.UFix64).String())
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
	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

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
			t.Fatalf("transfer_arena tx execution: %v", result.Error)
		}

		// Validate new balance
		bal := arenaBalance(t, em, newAcct)
		if bal != amount {
			t.Fatalf("Incorrect balance after minting, expected: %s, got: %s", amount, bal)
		}

		// check expected events
		validateEvents(t, result, []string{
			"TokensWithdrawn",
			"TokensDeposited",
		})

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

		// check expected events
		validateEvents(t, result, []string{
			"TokensWithdrawn",
			"TokensDeposited",
		})
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

func TestDestroyAdministrator(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	// Check that current admin can do admin tasks, i.e. create minter
	amount, _ := cadence.NewUFix64("1000.0")
	tx := txRenderer.MintTokens(em.ServiceAccount, amount)
	signers := emulator.TxSigners{
		Proposer:    em.ServiceAccount,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{em.ServiceAccount},
	}
	em.SignTx(signers, tx)
	result := em.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("Expected mint to succeed but did not: %v", result.Error)
	}

	// Destroy the Administrator resource
	tx = txRenderer.DestroyAdministrator()
	signers = emulator.TxSigners{
		Proposer:    em.ServiceAccount,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{em.ServiceAccount},
	}
	em.SignTx(signers, tx)
	result = em.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("DestroyAdministrator transaction execution: %v", result.Error)
	}

	// Ensure the destruction event is emitted
	if len(result.Events) != 1 {
		t.Fatalf("Expected destruction event to be emitted")
	}
	if !strings.Contains(result.Events[0].Type, "AdministratorDestroyed") {
		t.Fatalf("Expected AdministratorDestroyed event but got: %v", result.Events[0].Type)
	}

	// Old admin should not be able to mint
	tx = txRenderer.MintTokens(em.ServiceAccount, amount)
	signers = emulator.TxSigners{
		Proposer:    em.ServiceAccount,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{em.ServiceAccount},
	}
	em.SignTx(signers, tx)
	result = em.ExecuteTxWaitForSeal(tx)
	if result.Error == nil {
		t.Fatalf("Expected old admin mint to revert but did not")
	}

}

func TestTransferAdmininstrator(t *testing.T) {

	em, teardown := emulator.NewUnit(t, "3569", *dockerLogsOnFail)
	defer teardown()

	// Deploy ArenaToken contract to service account
	contractSource := arenatoken.Contract(em.Contracts["FungibleToken"])
	DeployContract(t, em, em.ServiceAccount, "ArenaToken", contractSource)
	txRenderer := arenatoken.New(em.Contracts["ArenaToken"], em.Contracts["FungibleToken"])

	// Check that current admin can do admin tasks, i.e. create minter
	amount, _ := cadence.NewUFix64("1000.0")
	tx := txRenderer.MintTokens(em.ServiceAccount, amount)
	signers := emulator.TxSigners{
		Proposer:    em.ServiceAccount,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{em.ServiceAccount},
	}
	em.SignTx(signers, tx)
	result := em.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("Expected mint to succeed but did not: %v", result.Error)
	}

	// Make a new account and ensure it can't do admin tasks
	newAcct := AddAccount(t, em)
	tx = txRenderer.MintTokens(em.ServiceAccount, amount)
	signers = emulator.TxSigners{
		Proposer:    newAcct,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{newAcct},
	}
	em.SignTx(signers, tx)
	result = em.ExecuteTxWaitForSeal(tx)
	if result.Error == nil {
		t.Fatalf("Expected non-admin mint to revert but did not")
	}

	// Transfer ownership of the Administrator resource to the new account
	tx = txRenderer.TransferAdministrator(em.ServiceAccount, newAcct)
	signers = emulator.TxSigners{
		Proposer:    em.ServiceAccount,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{em.ServiceAccount, newAcct},
	}
	em.SignTx(signers, tx)
	result = em.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("transfer_admin tx execution: %v", result.Error)
	}

	// New account should now be able to mint
	tx = txRenderer.MintTokens(em.ServiceAccount, amount)
	signers = emulator.TxSigners{
		Proposer:    newAcct,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{newAcct},
	}
	em.SignTx(signers, tx)
	result = em.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("Expected new admin to mint successfully")
	}

	// Old admin should not be able to mint
	tx = txRenderer.MintTokens(em.ServiceAccount, amount)
	signers = emulator.TxSigners{
		Proposer:    em.ServiceAccount,
		Payer:       em.ServiceAccount,
		Authorizers: []flow.Address{em.ServiceAccount},
	}
	em.SignTx(signers, tx)
	result = em.ExecuteTxWaitForSeal(tx)
	if result.Error == nil {
		t.Fatalf("Expected old admin mint to revert but did not")
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

func validateEvents(t *testing.T, result *flow.TransactionResult, expected []string) {

	if len(result.Events) != len(expected) {
		t.Fatalf("Unexpected number of events")
	}
	for i, e := range result.Events {
		if !strings.Contains(e.Type, expected[i]) {
			t.Fatalf("Expected event type: %s, got: %s", expected[i], e.Type)
		}
	}
}
