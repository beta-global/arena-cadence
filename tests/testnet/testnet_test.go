package testnet

import (
	"testing"

	"github.com/arena/arena-cadence/lib/go/arenatoken"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"
)

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

	})
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
