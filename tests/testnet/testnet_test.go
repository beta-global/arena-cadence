package testnet

import (
	"context"
	"testing"

	"github.com/arena/arena-cadence/lib/go/arenatoken"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

func TestAdminActions(t *testing.T) {

	// establish testnet connection and load necessary keys
	flowclient, err := client.New(testnetRPC, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Creating testnet client: %v", err)
	}
	adminAddr := flow.HexToAddress("0x0996b5100d5c8ad6")
	adminPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "8df3e8728ce2a9271107861a3c11c5def5c92120ba911ffff3beb5267e87680d")
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}
	keys := make(map[flow.Address]crypto.PrivateKey)
	keys[adminAddr] = adminPrivkey
	tc := &testnetClient{
		flowclient: flowclient,
		privkeys:   keys,
	}
	txRenderer := arenatoken.NewRenderer(adminAddr, fungibleTokenAddr)

	t.Run("MintToAdmin", func(t *testing.T) {
		// fetch the current balance
		oldBalance := arenaBalance(t, tc, adminAddr)

		// mint some tokens
		mintAmount, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.MintTokens(adminAddr, mintAmount)
		signers := txSigners{
			Proposer:    adminAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{adminAddr},
		}
		tc.SignTx(signers, tx)
		result := tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("mint_arena tx execution: %v", result.Error)
		}

		newBalance := arenaBalance(t, tc, adminAddr)

		// validate the new balance
		target := oldBalance + mintAmount
		if newBalance != target {
			t.Fatalf("Expected balance: %s, got: %s", target, newBalance)
		}
	})

	t.Run("MintToUser", func(t *testing.T) {
		userAddr := flow.HexToAddress("0x15b169c50310d253")

		// fetch the current balance
		oldBalance := arenaBalance(t, tc, userAddr)

		// mint some tokens
		mintAmount, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.MintTokens(userAddr, mintAmount)
		signers := txSigners{
			Proposer:    adminAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{adminAddr},
		}
		tc.SignTx(signers, tx)
		result := tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("mint_arena tx execution: %v", result.Error)
		}

		newBalance := arenaBalance(t, tc, userAddr)

		// validate the new balance
		target := oldBalance + mintAmount
		if newBalance != target {
			t.Fatalf("Expected balance: %s, got: %s", target, newBalance)
		}
	})

	t.Run("MintToUninitializedAccount", func(t *testing.T) {
		uninitializedUser := flow.HexToAddress("0xade1692b19cf30f1")

		// mint some tokens and expect tx to revert because the account is
		// not configured to recieve arena tokens
		mintAmount, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.MintTokens(uninitializedUser, mintAmount)
		signers := txSigners{
			Proposer:    adminAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{adminAddr},
		}
		tc.SignTx(signers, tx)
		result := tc.ExecuteTxWaitForSeal(tx)

		// fail if tx didn't revert
		if result.Error == nil {
			t.Fatalf("Expected mint to revert but did not: %v", result.Error)
		}
	})

}

func TestStandardUserFlow(t *testing.T) {

	// establish testnet connection and load necessary keys
	flowclient, err := client.New(testnetRPC, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Creating testnet client: %v", err)
	}

	adminAddr := flow.HexToAddress("0x0996b5100d5c8ad6")
	adminPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "8df3e8728ce2a9271107861a3c11c5def5c92120ba911ffff3beb5267e87680d")
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	userAddr := flow.HexToAddress("0x15b169c50310d253")
	userPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "a5d734436c43463019bc161294e22cad504d3d956a5d3e7f3a71989f83eaca44")
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	keys := make(map[flow.Address]crypto.PrivateKey)

	keys[adminAddr] = adminPrivkey
	keys[userAddr] = userPrivkey

	tc := testnetClient{
		flowclient: flowclient,
		privkeys:   keys,
	}
	txRenderer := arenatoken.NewRenderer(adminAddr, fungibleTokenAddr)

	t.Run("SetupAccount", func(t *testing.T) {

		// Run the account setup TX for the user account
		tx := txRenderer.SetupAccount()
		signers := txSigners{
			Proposer:    userAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{userAddr},
		}
		tc.SignTx(signers, tx)

		result := tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("setup_account tx execution: %v", result.Error)
		}

	})

	t.Run("TransferArena", func(t *testing.T) {

		// Send some arena to the new account from the admin
		amt, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.Transfer(userAddr, amt)
		signers := txSigners{
			Proposer:    adminAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{adminAddr},
		}
		tc.SignTx(signers, tx)

		result := tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("Admin to user transfer tx execution: %v", result.Error)
		}

		// Send the same amount back to the admin
		tx = txRenderer.Transfer(adminAddr, amt)
		signers = txSigners{
			Proposer:    userAddr,
			Payer:       userAddr,
			Authorizers: []flow.Address{userAddr},
		}
		tc.SignTx(signers, tx)

		result = tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("User to admin transfer tx execution: %v", result.Error)
		}

		// Try to send more that user balance and expect tx to revert
		amt, _ = cadence.NewUFix64("9999999999.0")
		tx = txRenderer.Transfer(adminAddr, amt)
		signers = txSigners{
			Proposer:    userAddr,
			Payer:       userAddr,
			Authorizers: []flow.Address{userAddr},
		}
		tc.SignTx(signers, tx)

		result = tc.ExecuteTxWaitForSeal(tx)
		if result.Error == nil {
			t.Fatalf("Expected transfer to revert but did not: %v", result.Error)
		}

	})

}

func TestStandardUserFlowAdminPaysFees(t *testing.T) {

	// establish testnet connection and load necessary keys
	flowclient, err := client.New(testnetRPC, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Creating testnet client: %v", err)
	}

	adminAddr := flow.HexToAddress("0x0996b5100d5c8ad6")
	adminPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "8df3e8728ce2a9271107861a3c11c5def5c92120ba911ffff3beb5267e87680d")
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	userAddr := flow.HexToAddress("0x15b169c50310d253")
	userPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "a5d734436c43463019bc161294e22cad504d3d956a5d3e7f3a71989f83eaca44")
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	keys := make(map[flow.Address]crypto.PrivateKey)

	keys[adminAddr] = adminPrivkey
	keys[userAddr] = userPrivkey

	tc := testnetClient{
		flowclient: flowclient,
		privkeys:   keys,
	}
	txRenderer := arenatoken.NewRenderer(adminAddr, fungibleTokenAddr)

	t.Run("SetupAccount", func(t *testing.T) {

		// Run the account setup TX for the user account
		tx := txRenderer.SetupAccount()
		signers := txSigners{
			Proposer:    userAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{userAddr},
		}
		tc.SignTx(signers, tx)

		result := tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("setup_account tx execution: %v", result.Error)
		}

	})

	t.Run("TransferArena", func(t *testing.T) {

		// Send some arena to the new account from the admin
		amt, _ := cadence.NewUFix64("100.0")
		tx := txRenderer.Transfer(userAddr, amt)
		signers := txSigners{
			Proposer:    adminAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{adminAddr},
		}
		tc.SignTx(signers, tx)

		result := tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("Admin to user transfer tx execution: %v", result.Error)
		}

		// Send the same amount back to the admin
		tx = txRenderer.Transfer(adminAddr, amt)
		signers = txSigners{
			Proposer:    userAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{userAddr},
		}
		tc.SignTx(signers, tx)

		result = tc.ExecuteTxWaitForSeal(tx)
		if result.Error != nil {
			t.Fatalf("User to admin transfer tx execution: %v", result.Error)
		}

		// Try to send more that user balance and expect tx to revert
		amt, _ = cadence.NewUFix64("9999999999.0")
		tx = txRenderer.Transfer(adminAddr, amt)
		signers = txSigners{
			Proposer:    userAddr,
			Payer:       adminAddr,
			Authorizers: []flow.Address{userAddr},
		}
		tc.SignTx(signers, tx)

		result = tc.ExecuteTxWaitForSeal(tx)
		if result.Error == nil {
			t.Fatalf("Expected transfer to revert but did not: %v", result.Error)
		}

	})

}

func arenaBalance(t *testing.T, tc *testnetClient, target flow.Address) cadence.UFix64 {
	t.Helper()

	txRenderer := arenatoken.NewRenderer(arenaTokenAddr, fungibleTokenAddr)
	balanceScript, args := txRenderer.Balance(target)
	val, err := tc.flowclient.ExecuteScriptAtLatestBlock(context.Background(), balanceScript, args)
	if err != nil {
		t.Fatalf("Reading balance: %v", err)
	}

	return val.(cadence.UFix64)
}

/*
func TestDeploy(t *testing.T) {

	flowclient, err := client.New(testnetRPC, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Creating testnet client: %v", err)
	}
	testnetAddr := flow.HexToAddress("0x0996b5100d5c8ad6")
	testnetPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "8df3e8728ce2a9271107861a3c11c5def5c92120ba911ffff3beb5267e87680d")
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}
	keys := make(map[flow.Address]crypto.PrivateKey)
	keys[testnetAddr] = testnetPrivkey

	tc := testnetClient{
		flowclient: flowclient,
		privkeys:   keys,
	}

	contractSource := arenatoken.Contract(flow.HexToAddress(fungibleTokenAddr))
	contractName := "ArenaToken"

	// Deploy the contract to a testnet account
	const deployContractTemplate = `
	transaction(name: String, code: String) {
		prepare(signer: AuthAccount) {
			signer.contracts.add(name: name, code: code.decodeHex())
		}
	}`

	tx := flow.NewTransaction().
		SetScript([]byte(deployContractTemplate)).
		AddRawArgument(jsoncdc.MustEncode(cadence.NewString(contractName))).
		AddRawArgument(jsoncdc.MustEncode(cadence.NewString(hex.EncodeToString([]byte(contractSource)))))

	signers := txSigners{
		Proposer:    testnetAddr,
		Authorizers: []flow.Address{testnetAddr},
		Payer:       testnetAddr,
	}
	if err := tc.SignTx(signers, tx); err != nil {
		t.Fatalf("Failed to sign Tx: %v", err)
	}

	result := tc.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		t.Fatalf("Executing contract deploy: %v", err)
	}

	spew.Dump(result)
}
*/
