{{ import "ArenaToken" }}
{{ import "FungibleToken" }}

pub fun main(account: Address): UFix64 {

    let ArenaReceiverRef = getAccount(account)
        .getCapability(/public/arenaTokenBalance)
        .borrow<&ArenaToken.Vault{FungibleToken.Balance}>()
        ?? panic("Unable to borrow reference to Arena vault")

    log("Account 0x{{ .Address }} Balance")
    log(ArenaReceiverRef.balance)

    return ArenaReceiverRef.balance
}`
