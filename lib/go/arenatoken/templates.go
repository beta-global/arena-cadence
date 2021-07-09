package arenatoken

import (
	"log"

	arenacadence "github.com/arena/arena-cadence"
)

var (
	contractTemplate              string
	setupAccountTemplate          string
	mintArenaTemplate             string
	balanceTemplate               string
	transferTemplate              string
	transferAdministratorTemplate string
	destroyAdministratorTemplate  string
	burnTemplate                  string
)

// read templates from embedded fs
func init() {
	// contracts
	contractTemplate = readTemplate("cadence/contracts/arenatoken.cdc")

	// transactions
	setupAccountTemplate = readTemplate("cadence/transactions/arenaToken/setup_account.cdc")
	mintArenaTemplate = readTemplate("cadence/transactions/arenaToken/mint_arena.cdc")
	destroyAdministratorTemplate = readTemplate("cadence/transactions/arenaToken/destroy_admin.cdc")
	transferTemplate = readTemplate("cadence/transactions/arenaToken/transfer.cdc")
	transferAdministratorTemplate = readTemplate("cadence/transactions/arenaToken/transfer_admin.cdc")
	burnTemplate = readTemplate("cadence/transactions/arenaToken/burn_arena.cdc")

	// scripts
	balanceTemplate = readTemplate("cadence/scripts/arenaToken/balance.cdc")
}

func readTemplate(path string) string {
	tpl, err := arenacadence.Cadence.ReadFile(path)
	if err != nil {
		log.Fatalf("Missing embedded template: %v", err)
	}
	return string(tpl)
}
