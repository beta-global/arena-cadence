import * as fcl from "@onflow/fcl";

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

export {Transaction, Script}