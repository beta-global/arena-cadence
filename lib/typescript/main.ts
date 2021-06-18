import { ArenaTokenService } from "./arena-token";
console.log("hello")

const arenaTokenService = new ArenaTokenService(
	"0xABC",
	"0xDEF"
);

const tx = arenaTokenService.sendArena("0x123", 100);
console.log(tx)
console.log("done")

const getBalance = arenaTokenService.getBalance("0xABCDEF")
console.log(getBalance)

const setupAccount = arenaTokenService.setupAccount("0xABC123")
console.log(setupAccount)
