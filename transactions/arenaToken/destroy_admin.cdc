{{ import "ArenaToken" }}

transaction() {

    prepare(currentAdmin: AuthAccount) {

        // Fetch current Administrator resource and replace it with nil
        let oldAdmin <- currentAdmin.load<@ArenaToken.Administrator>(from: ArenaToken.AdminStoragePath)!

        // Destroy the Administrator resource
        destroy oldAdmin
    }

    execute {}
}
