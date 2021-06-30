package arenatoken

import (
	"log"

	arenacadence "github.com/arena/arena-cadence"
)

var (
	contractTemplate     string
	setupAccountTemplate string
)

// read templates from embedded fs
func init() {
	contractTemplate = readContractTemplate("contracts/arenatoken.cdc")
	setupAccountTemplate = readTxTemplate("transactions/arenaToken/setup_account.cdc")
}

func readContractTemplate(path string) string {
	tpl, err := arenacadence.Contracts.ReadFile(path)
	if err != nil {
		log.Fatalf("Missing embedded template: %v", err)
	}
	return string(tpl)
}

func readTxTemplate(path string) string {
	tpl, err := arenacadence.Transactions.ReadFile(path)
	if err != nil {
		log.Fatalf("Missing embedded template: %v", err)
	}
	return string(tpl)
}
