{{ import "ArenaToken" }}
{{ import "FungibleToken" }}

pub fun main(account: Address): UFix64 {

    let ArenaBalanceRef = getAccount(account)
        .getCapability(ArenaToken.BalancePublicPath)!
        .borrow<&ArenaToken.Vault{FungibleToken.Balance}>()
        ?? panic("Unable to borrow reference to Arena vault")

    return ArenaBalanceRef.balance
}
