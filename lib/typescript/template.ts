import * as fcl from "@onflow/fcl";
import * as fs from "fs";
import * as path from "path"

type Transaction = {
	name: string
	code: string
	gasLimit: Number
	args: fcl.arg[]
}

type Script = {
	name: string
	code: string
	args: fcl.arg[]
}

function resolveImports(tpl: string, mappings: Map<string, string>): string {
	for (let [contract, address] of mappings) {
		tpl = tpl.replace(`{{ import "${contract}" }}`, `import "${contract}" from ${fcl.withPrefix(address)}`)
	}
	return tpl
}

function readTemplate(tplpath: string): string {
	return fs.readFileSync(
			path.join(__dirname, `../../../${tplpath}`),
			"utf8"
		)
}

export {Transaction, Script, resolveImports, readTemplate}