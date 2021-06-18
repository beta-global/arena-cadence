// @ts-ignore
import * as fcl from "@onflow/fcl";
import * as t from "@onflow/types";
import * as fs from "fs";
import * as path from "path";

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

class ArenaTokenService {
	constructor(
		private readonly fungibleTokenAddress: string,
		private readonly arenaTokenAddress: string,
	) {}

	sendArena(recipient: string, amount: number): Transaction {
		const template = readTemplate("transactions/arenaToken/send_arena.cdc")
		const code = resolveImports(template, new Map([
			["ArenaToken", this.arenaTokenAddress],
			["FungibleToken", this.fungibleTokenAddress]
		]));
		
		return {
			name: "Send Arena",
			code: code,
			args: [
				fcl.arg(recipient, t.Address),
				fcl.arg(amount.toFixed(8).toString(), t.UFix64)
			],
			gasLimit: 25
		}
	}
	
	setupAccount(recipient: string): Transaction {
		const template = readTemplate("transactions/arenaToken/setup_account.cdc")
		const code = resolveImports(template, new Map([
			["ArenaToken", this.arenaTokenAddress],
			["FungibleToken", this.fungibleTokenAddress]
		]));
		
		return {
			name: "Send Arena",
			code: code,
			args: [
				fcl.arg(recipient, t.Address),
			],
			gasLimit: 50
		}
	}
	
	getBalance(target: string): Script {
		const template = readTemplate("scripts/arenaToken/get_balance.cdc")
		const code = resolveImports(template, new Map([
			["ArenaToken", this.arenaTokenAddress],
			["FungibleToken", this.fungibleTokenAddress]
		]));
		
		return {
			name: "Send Arena",
			code: code,
			args: [
				fcl.arg(target, t.Address),
			],
		}
	}

}

export { ArenaTokenService };
