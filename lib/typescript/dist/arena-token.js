"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ArenaTokenTemplates = void 0;
var fcl = __importStar(require("@onflow/fcl"));
var t = __importStar(require("@onflow/types"));
var template_1 = require("./template");
var ArenaTokenTemplates = /** @class */ (function () {
    function ArenaTokenTemplates(fungibleTokenAddress, arenaTokenAddress) {
        this.fungibleTokenAddress = fungibleTokenAddress;
        this.arenaTokenAddress = arenaTokenAddress;
    }
    ArenaTokenTemplates.prototype.sendArena = function (recipient, amount) {
        var template = template_1.readTemplate("transactions/arenaToken/send_arena.cdc");
        var code = template_1.resolveImports(template, new Map([
            ["ArenaToken", this.arenaTokenAddress],
            ["FungibleToken", this.fungibleTokenAddress]
        ]));
        return {
            name: "SendArena",
            code: code,
            args: [
                fcl.arg(recipient, t.Address),
                fcl.arg(amount.toFixed(8).toString(), t.UFix64)
            ],
            gasLimit: 25
        };
    };
    ArenaTokenTemplates.prototype.setupAccount = function (recipient) {
        var template = template_1.readTemplate("transactions/arenaToken/setup_account.cdc");
        var code = template_1.resolveImports(template, new Map([
            ["ArenaToken", this.arenaTokenAddress],
            ["FungibleToken", this.fungibleTokenAddress]
        ]));
        return {
            name: "SetupAccount",
            code: code,
            args: [
                fcl.arg(recipient, t.Address),
            ],
            gasLimit: 50
        };
    };
    ArenaTokenTemplates.prototype.getBalance = function (target) {
        var template = template_1.readTemplate("scripts/arenaToken/get_balance.cdc");
        var code = template_1.resolveImports(template, new Map([
            ["ArenaToken", this.arenaTokenAddress],
            ["FungibleToken", this.fungibleTokenAddress]
        ]));
        return {
            name: "GetBalance",
            code: code,
            args: [
                fcl.arg(target, t.Address),
            ],
        };
    };
    return ArenaTokenTemplates;
}());
exports.ArenaTokenTemplates = ArenaTokenTemplates;
