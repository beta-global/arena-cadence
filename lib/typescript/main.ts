import { ArenaTokenTemplates } from "./arena-token";

const arenaTokenService = new ArenaTokenTemplates(
	"0xf8d6e0586b0a20c7",
	"0x01cf0e2f2f715450"
);

const tx = arenaTokenService.sendArena("0x123", 100);
console.log(tx)
console.log("done")

const getBalance = arenaTokenService.getBalance("0xABCDEF")
console.log(getBalance)

const setupAccount = arenaTokenService.setupAccount("0xABC123")
console.log(setupAccount)
