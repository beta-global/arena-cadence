package emulator

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	jsoncdc "github.com/onflow/cadence/encoding/json"

	"github.com/arena/arena-cadence/tests/docker"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/crypto"
	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

var (
	DefaultImage       = "gcr.io/flow-container-registry/emulator:0.19.0"
	DefaultPort        = "3569"
	ServiceAccountKey  = "2eae2f31cb5b756151fa11d82949c634b8f28796a711d7eb1e52cc301ed11111"
	ServiceAccountAddr = "f8d6e0586b0a20c7"
	// Well known contracts
	FungibleTokenAddr = "ee82856bf20e2aa6"
	FlowTokenAddr     = "0ae53cb6e3f42a79"
)

type EmulatorContainer struct {
	Image string
	Port  string
	Args  []string
}

type Emulator struct {
	Client         *client.Client
	Privkeys       map[flow.Address]crypto.PrivateKey
	Contracts      map[string]flow.Address
	ServiceAccount flow.Address
}

// NewUnit starts an instance of the flow emulator in a docker container and
// returns a teardown function that should be invoked after the test is complete.
// The emulator has no initial state other than several base flow contracts.
func NewUnit(t *testing.T, port string, dockerLogsOnFail bool) (em *Emulator, teardown func()) {

	// start emulator container
	// TODO(dave): make port injectable so we can run tests in parallel
	c := docker.StartContainer(t, DefaultImage, DefaultPort,
		"-p", fmt.Sprintf("%s:%s", port, port),
		"-e", fmt.Sprintf("FLOW_PORT=%s", port),
		"-e", "FLOW_VERBOSE=true",
		"-e", "FLOW_SERVICEPUBLICKEY=31a053a2003d95760d8fff623aeedcc927022d8e0767972ab507608a5f611636e81857c6c46b048be6f66eddc13f5553627861153f6ce301caf5a056d68efc29",
		"-e", "FLOW_SERVICEKEYSIGALGO=ECDSA_P256",
		"-e", "FLOW_SERVICEKEYHASHALGO=SHA3_256",
	)

	client, err := client.New(fmt.Sprintf(":%s", port), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Opening rpc connection: %v", err)
	}

	t.Log("Establishing connection to emulator ...")
	var success bool
	for tries := 15; tries > 0; tries-- {
		if err := client.Ping(context.Background()); err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		success = true
		break
	}
	if !success {
		if dockerLogsOnFail {
			docker.DumpContainerLogs(t, c.ID)
		}
		docker.StopContainer(t, c.ID)
		t.Fatalf("Unable to connect to emulator")
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown = func() {
		t.Helper()

		// Dump container logs if the test failed
		if t.Failed() && dockerLogsOnFail {
			docker.DumpContainerLogs(t, c.ID)
		}

		docker.StopContainer(t, c.ID)
	}

	// Add service account key and known contracts
	// TODO(dave): figure out how to inject flow.json to container
	privkeys := make(map[flow.Address]crypto.PrivateKey)
	acctKey, _ := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, ServiceAccountKey)
	privkeys[flow.HexToAddress(ServiceAccountAddr)] = acctKey

	contracts := make(map[string]flow.Address)
	contracts["FungibleToken"] = flow.HexToAddress(FungibleTokenAddr)
	contracts["FlowToken"] = flow.HexToAddress(FlowTokenAddr)

	em = &Emulator{
		Client:         client,
		Privkeys:       privkeys,
		Contracts:      contracts,
		ServiceAccount: flow.HexToAddress(ServiceAccountAddr),
	}
	return em, teardown
}

type TxSigners struct {
	Proposer    flow.Address
	Payer       flow.Address
	Authorizers []flow.Address
}

func (e *Emulator) SignTx(signers TxSigners, tx *flow.Transaction) error {
	// TODO(dave): this is kinda ugly, but payer is part of the payload, so we can't
	// defer knowing the payer until the end. Try to think of a clean way to allow
	// signing to be idempotent. Might be able to pass id upfront but also provide
	// future resolver for signature similar to wallet spec authorization function impl

	if signers.Payer == flow.EmptyAddress {
		return fmt.Errorf("Tx payer signer must be specified")
	}

	// TODO(dave): will need to slightly adjust payer logic if we ever used
	// keys with fractional signing power
	// Nothing to do if envelope already signed
	if len(tx.EnvelopeSignatures) > 0 {
		return nil
	}

	// Get current block
	referenceBlockID, err := e.Client.GetLatestBlock(context.Background(), true)
	if err != nil {
		return fmt.Errorf("GetLatestBlock: %v", err)
	}

	// get updated sequence number for proposal key
	proposerAcct, err := e.Client.GetAccount(context.Background(), signers.Proposer)
	if err != nil {
		return fmt.Errorf("GetAccount: %v", err)
	}

	// finalize payload
	// TODO(dave): inejectable key_id for each signer role
	tx.SetPayer(signers.Payer)
	tx.SetProposalKey(signers.Proposer, proposerAcct.Keys[0].Index, proposerAcct.Keys[0].SequenceNumber)
	tx.SetReferenceBlockID(referenceBlockID.ID)
	for _, authorizer := range signers.Authorizers {
		tx.AddAuthorizer(authorizer)
	}

	// sign payload with proposal key if different than payer
	if signers.Proposer != flow.EmptyAddress && signers.Proposer != signers.Payer {

		// fetch signing key and sign
		signingKey := e.Privkeys[signers.Proposer]
		proposerSigner := crypto.NewInMemorySigner(signingKey, crypto.SHA3_256)
		if err := tx.SignPayload(signers.Proposer, proposerAcct.Keys[0].Index, proposerSigner); err != nil {
			return fmt.Errorf("Unable to sign payload with proposer key: %v", err)
		}
	}

	// sign payload with each authorizer key different than payer
	for _, authorizer := range signers.Authorizers {
		if authorizer != flow.EmptyAddress && authorizer == signers.Payer {
			continue
		}

		authorizerAcct, err := e.Client.GetAccount(context.Background(), authorizer)
		if err != nil {
			return fmt.Errorf("Unable to get account: %v", err)
		}

		// fetch signing key and sign
		signingKey := e.Privkeys[authorizer]
		authSigner := crypto.NewInMemorySigner(signingKey, crypto.SHA3_256)
		if err := tx.SignPayload(authorizer, authorizerAcct.Keys[0].Index, authSigner); err != nil {
			return fmt.Errorf("Unable to sign payload with authorizer key: %v", err)
		}
	}

	// Sign envelope with payer
	payerAcct, err := e.Client.GetAccount(context.Background(), signers.Payer)
	if err != nil {
		return fmt.Errorf("Unable to get Payer account: %v", err)
	}

	// fetch signing key and sign
	signingKey := e.Privkeys[signers.Payer]
	payerSigner := crypto.NewInMemorySigner(signingKey, crypto.SHA3_256)
	if err := tx.SignEnvelope(signers.Payer, payerAcct.Keys[0].Index, payerSigner); err != nil {
		return fmt.Errorf("Unable to sign payload with Payer key: %v", err)
	}

	return nil
}

func (e *Emulator) ExecuteTxWaitForSeal(tx *flow.Transaction) *flow.TransactionResult {
	if err := e.Client.SendTransaction(context.Background(), *tx); err != nil {
		return &flow.TransactionResult{Error: fmt.Errorf("Sending Tx: %v", err)}
	}

	result, err := e.Client.GetTransactionResult(context.Background(), tx.ID())
	if err != nil {
		return &flow.TransactionResult{Error: fmt.Errorf("GetTransactionResult: %v", err)}
	}

	for result.Status != flow.TransactionStatusSealed {
		fmt.Println("Waiting for tx to be sealed...")
		time.Sleep(5 * time.Second)
		result, err = e.Client.GetTransactionResult(context.Background(), tx.ID())
		if err != nil {
			return &flow.TransactionResult{Error: fmt.Errorf("GetTransactionResult: %v", err)}
		}
	}

	return result

}

const deployContractTemplate = `
transaction(name: String, code: String) {
	prepare(signer: AuthAccount) {
		signer.contracts.add(name: name, code: code.decodeHex())
	}
}`

func (e *Emulator) DeployContract(owner flow.Address, name string, source string) (*flow.TransactionResult, error) {
	tx := flow.NewTransaction().
		SetScript([]byte(deployContractTemplate)).
		AddRawArgument(jsoncdc.MustEncode(cadence.NewString(name))).
		AddRawArgument(jsoncdc.MustEncode(cadence.NewString(hex.EncodeToString([]byte(source)))))

	serviceAcct := flow.HexToAddress(ServiceAccountAddr)
	signers := TxSigners{
		Proposer:    serviceAcct,
		Authorizers: []flow.Address{owner},
		Payer:       serviceAcct,
	}
	if err := e.SignTx(signers, tx); err != nil {
		return nil, fmt.Errorf("Failed to sign Tx: %v", err)
	}

	result := e.ExecuteTxWaitForSeal(tx)

	// Add to mapping of tracked contracts if deploy succeeded
	if result.Error == nil {
		e.Contracts[name] = owner
	}
	return result, result.Error
}

const createAccountTemplate = `
transaction(publicKeys: [String], contracts: {String: String}) {
       prepare(signer: AuthAccount) {
               let acct = AuthAccount(payer: signer)

               for key in publicKeys {
                       acct.addPublicKey(key.decodeHex())
               }

               for contract in contracts.keys {
                       acct.contracts.add(name: contract, code: contracts[contract]!.decodeHex())
               }
       }
}
`

// AddAccount creates a new flow account utilizing a new randomly generated key.
// The private key is tracked by the emulator to facilitate signing transactions.
func (e *Emulator) AddAccount() (flow.Address, error) {

	// random seed
	rand.Seed(time.Now().UnixNano())
	randBuf := make([]byte, 32)
	rand.Read(randBuf)

	privkey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, randBuf)
	if err != nil {
		return flow.EmptyAddress, fmt.Errorf("Unable to create private key from seed: %v", err)
	}

	// construct an account key from the public key
	pubkey := privkey.PublicKey()
	accountKey := flow.NewAccountKey().
		SetPublicKey(pubkey).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(flow.AccountKeyWeightThreshold)
	pubkeys := make([]cadence.Value, 1)
	pubkeys[0] = cadence.NewString(hex.EncodeToString(accountKey.Encode()))

	// Convert to cadence specific format
	cadencePublicKeys := cadence.NewArray(pubkeys)
	cadenceContracts := cadence.NewDictionary(make([]cadence.KeyValuePair, 0))

	tx := flow.NewTransaction().
		SetScript([]byte(createAccountTemplate)).
		AddRawArgument(jsoncdc.MustEncode(cadencePublicKeys)).
		AddRawArgument(jsoncdc.MustEncode(cadenceContracts))

	serviceAcct := e.ServiceAccount
	signers := TxSigners{
		Proposer:    serviceAcct,
		Payer:       serviceAcct,
		Authorizers: []flow.Address{serviceAcct},
	}
	e.SignTx(signers, tx)

	result := e.ExecuteTxWaitForSeal(tx)
	if result.Error != nil {
		return flow.EmptyAddress, fmt.Errorf("Unable to create new account: %v", result.Error)
	}
	accountCreatedEvent := flow.AccountCreatedEvent(result.Events[0])
	newAcctAddr := accountCreatedEvent.Address()

	// Track the key of new account to simplify testing
	e.Privkeys[newAcctAddr] = privkey

	return newAcctAddr, nil
}
