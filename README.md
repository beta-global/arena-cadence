# arena-cadence
Cadence contracts and client libraries for creating transactions to interact with Arena flow blockchain resources. For the purposes of mainnet review, this repo only contains cadence code for the `ArenaToken` fungible token (code for additional resources/contracts will be submitted separately at a later date). All cadence code lives within the `cadence` directory.  

## Testing ##
  
  ```
  // Run tests against docker emulator
  go test ./tests -v
  
  // Run tests against sample testnet deployment
  go test ./tests/testnet -v
  ```
  
## Sample Usage ##

  ``` 
userAddr := flow.HexToAddress("0x15b169c50310d253")
userPrivkey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, sampleUserPrivekey)
if err != nil {
     log.Fatalf("Failed to decode private key: %v", err)
}

txRenderer := arenatoken.New(contractAddr, fungibleTokenAddr)

// Run the account setup TX for the user account
tx := txRenderer.SetupAccount()
signers := txSigners{
     Proposer:    userAddr,
     Payer:       userAddr,
     Authorizers: []flow.Address{userAddr},
}
SignTx(signers, tx, userPrivKey)

result := ExecuteTxWaitForSeal(tx)
if result.Error != nil {
      log.Fatalf("setup_account tx execution: %v", result.Error)
}
  ```
