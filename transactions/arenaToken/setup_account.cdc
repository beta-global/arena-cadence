// This transaction is a template for a transaction
// to add a Vault resource to their account
// so that they can use the ArenaToken

{{ import "FungibleToken" }} 
{{ import "ArenaToken" }}

transaction {

    var addr: Address

    prepare(signer: AuthAccount) {
        self.addr = signer.address

        //  Return early if the account already stores a ArenaToken Vault
        if signer.borrow<&ArenaToken.Vault>(from: /storage/arenaTokenVault) != nil {
            return
        }

        // Create a new ArenaToken Vault and put it in storage
        signer.save(
            <-ArenaToken.createEmptyVault(),
            to: /storage/arenaTokenVault
        )

        // Create a public capability to the Vault that only exposes
        // the deposit function through the Receiver interface
        signer.link<&ArenaToken.Vault{FungibleToken.Receiver}>(
            /public/arenaTokenReceiver,
            target: /storage/arenaTokenVault
        )

        // Create a public capability to the Vault that only exposes
        // the balance field through the Balance interface
        signer.link<&ArenaToken.Vault{FungibleToken.Balance}>(
            /public/arenaTokenBalance,
            target: /storage/arenaTokenVault
        )

    }

    post {

        getAccount(self.addr).getCapability(/public/arenaTokenReceiver)
            .check<&ArenaToken.Vault{FungibleToken.Receiver}>():
                "Receiver capability not created correctly"

        getAccount(self.addr).getCapability(/public/arenaTokenBalance)
            .check<&ArenaToken.Vault{FungibleToken.Balance}>():
                "Balance capability not created correctly"
    }
}`

