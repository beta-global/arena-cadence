// Package arenacadence statically embeds all the various arena cadence templates.
// go:embed does not support `..` in filepaths so this package exists as a compromise
// to statically embed the cadence templates while still conforming to the officially
// recommended cadence project structure https://joshuahannan.medium.com/how-i-organize-my-cadence-projects-75b811b700d9
package arenacadence

import (
	"embed"
)

// Statically embed all cadence templates within the module
//go:embed cadence
var Cadence embed.FS
