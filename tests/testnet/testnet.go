package testnet

import (
	"context"
	"fmt"
	"time"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

var (
	fungibleTokenAddr = flow.HexToAddress("0x9a0766d93b6608b7")
	testnetRPC        = "access.devnet.nodes.onflow.org:9000"
)

type testnetClient struct {
	flowclient *client.Client
	privkeys   map[flow.Address]crypto.PrivateKey
}

type txSigners struct {
	Proposer    flow.Address
	Payer       flow.Address
	Authorizers []flow.Address
}

func (c *testnetClient) SignTx(signers txSigners, tx *flow.Transaction) error {

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
	referenceBlockID, err := c.flowclient.GetLatestBlock(context.Background(), true)
	if err != nil {
		return fmt.Errorf("GetLatestBlock: %v", err)
	}

	// get updated sequence number for proposal key
	proposerAcct, err := c.flowclient.GetAccount(context.Background(), signers.Proposer)
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
		signingKey := c.privkeys[signers.Proposer]
		proposerSigner := crypto.NewInMemorySigner(signingKey, crypto.SHA3_256)
		if err := tx.SignPayload(signers.Proposer, proposerAcct.Keys[0].Index, proposerSigner); err != nil {
			return fmt.Errorf("Unable to sign payload with proposer key: %v", err)
		}
	}

	// sign payload with each authorizer key different than payer
	for _, authorizer := range signers.Authorizers {
		// Don't need to sign if sig will be part of the envelope
		if authorizer != flow.EmptyAddress && authorizer == signers.Payer {
			continue
		}

		// Don't need to sign if already added a payload signature as a proposer
		if authorizer != flow.EmptyAddress && authorizer == signers.Proposer {
			continue
		}

		authorizerAcct, err := c.flowclient.GetAccount(context.Background(), authorizer)
		if err != nil {
			return fmt.Errorf("Unable to get account: %v", err)
		}

		// fetch signing key and sign
		signingKey := c.privkeys[authorizer]
		authSigner := crypto.NewInMemorySigner(signingKey, crypto.SHA3_256)
		if err := tx.SignPayload(authorizer, authorizerAcct.Keys[0].Index, authSigner); err != nil {
			return fmt.Errorf("Unable to sign payload with authorizer key: %v", err)
		}
	}

	// Sign envelope with payer
	payerAcct, err := c.flowclient.GetAccount(context.Background(), signers.Payer)
	if err != nil {
		return fmt.Errorf("Unable to get Payer account: %v", err)
	}

	// fetch signing key and sign
	signingKey := c.privkeys[signers.Payer]
	payerSigner := crypto.NewInMemorySigner(signingKey, crypto.SHA3_256)
	if err := tx.SignEnvelope(signers.Payer, payerAcct.Keys[0].Index, payerSigner); err != nil {
		return fmt.Errorf("Unable to sign payload with Payer key: %v", err)
	}

	return nil
}

func (c *testnetClient) ExecuteTxWaitForSeal(tx *flow.Transaction) *flow.TransactionResult {
	if err := c.flowclient.SendTransaction(context.Background(), *tx); err != nil {
		return &flow.TransactionResult{Error: fmt.Errorf("Sending Tx: %v", err)}
	}

	result, err := c.flowclient.GetTransactionResult(context.Background(), tx.ID())
	if err != nil {
		return &flow.TransactionResult{Error: fmt.Errorf("GetTransactionResult: %v", err)}
	}

	for result.Status != flow.TransactionStatusSealed {
		fmt.Println("Waiting for tx to be sealed...")
		time.Sleep(5 * time.Second)
		result, err = c.flowclient.GetTransactionResult(context.Background(), tx.ID())
		if err != nil {
			return &flow.TransactionResult{Error: fmt.Errorf("GetTransactionResult: %v", err)}
		}
	}

	return result
}
