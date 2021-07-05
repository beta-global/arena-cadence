{{ import "ArenaToken" }}

transaction() {

    prepare(currentAdmin: AuthAccount, newAdmin: AuthAccount) {

        // Retrieve the admin object from storage of existing admin
        let admin <- currentAdmin.load<@ArenaToken.Administrator>(from: ArenaToken.AdminStoragePath)!

        newAdmin.save(
            <-admin,
            to: ArenaToken.AdminStoragePath
        )
    }

    execute {}
}
