{{ import "FungibleToken" }}
{{ import "ArenaToken" }}

transaction(amount: UFix64) {
    let tokenAdmin: &ArenaToken.Administrator
    let burnVault: @FungibleToken.Vault

    prepare(admin: AuthAccount) {
        self.tokenAdmin = admin.borrow<&ArenaToken.Administrator>(from: ArenaToken.AdminStoragePath)
            ?? panic("Signer is not the token admin")

        // Withdraw the amount we intend to burn
        let vaultRef = admin.borrow<&ArenaToken.Vault>(from: ArenaToken.VaultStoragePath)
            ?? panic("Could not borrow reference to the admin's Vault!")

        self.burnVault <- vaultRef.withdraw(amount: amount)
    }

    execute {
        let burner <- self.tokenAdmin.createNewBurner()
        burner.burnTokens(from: <-self.burnVault)

        destroy burner
    }
}
