package arenatoken

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/onflow/flow-go-sdk"
)

// render executes the provided go template file resolving the provided contract imports
func render(tpl string, obj interface{}, contracts map[string]flow.Address) string {

	// capture contracts mapping via closure
	importExpander := func(c string) string {
		if _, ok := contracts[c]; !ok {
			return fmt.Sprintf("INVALID_IMPORT: %s", c)
		}
		return fmt.Sprintf("import %s from 0x%s", c, contracts[c])
	}
	funcMap := template.FuncMap{
		"import": importExpander,
	}

	// Create a template, add the function map, and parse the text.
	tmpl := template.Must(template.New("titleTest").Funcs(funcMap).Parse(tpl))

	// Run the template to verify the output.
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, obj)
	if err != nil {
		log.Fatalf("execution: %s", err)
	}

	return buf.String()

}
