// Package arenacadence statically embeds all the various arena cadence templates
package arenacadence

import (
	"embed"
)

// Statically embed all cadence templates within the module

//go:embed contracts
var Contracts embed.FS

//go:embed transactions
var Transactions embed.FS

//go:embed scripts
var Scripts embed.FS
