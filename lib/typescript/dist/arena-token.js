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
var __values = (this && this.__values) || function(o) {
    var s = typeof Symbol === "function" && Symbol.iterator, m = s && o[s], i = 0;
    if (m) return m.call(o);
    if (o && typeof o.length === "number") return {
        next: function () {
            if (o && i >= o.length) o = void 0;
            return { value: o && o[i++], done: !o };
        }
    };
    throw new TypeError(s ? "Object is not iterable." : "Symbol.iterator is not defined.");
};
var __read = (this && this.__read) || function (o, n) {
    var m = typeof Symbol === "function" && o[Symbol.iterator];
    if (!m) return o;
    var i = m.call(o), r, ar = [], e;
    try {
        while ((n === void 0 || n-- > 0) && !(r = i.next()).done) ar.push(r.value);
    }
    catch (error) { e = { error: error }; }
    finally {
        try {
            if (r && !r.done && (m = i["return"])) m.call(i);
        }
        finally { if (e) throw e.error; }
    }
    return ar;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ArenaTokenTemplates = void 0;
var fcl = __importStar(require("@onflow/fcl"));
var t = __importStar(require("@onflow/types"));
var fs = __importStar(require("fs"));
var path = __importStar(require("path"));
function resolveImports(tpl, mappings) {
    var e_1, _a;
    try {
        for (var mappings_1 = __values(mappings), mappings_1_1 = mappings_1.next(); !mappings_1_1.done; mappings_1_1 = mappings_1.next()) {
            var _b = __read(mappings_1_1.value, 2), contract = _b[0], address = _b[1];
            tpl = tpl.replace("{{ import \"" + contract + "\" }}", "import \"" + contract + "\" from " + fcl.withPrefix(address));
        }
    }
    catch (e_1_1) { e_1 = { error: e_1_1 }; }
    finally {
        try {
            if (mappings_1_1 && !mappings_1_1.done && (_a = mappings_1.return)) _a.call(mappings_1);
        }
        finally { if (e_1) throw e_1.error; }
    }
    return tpl;
}
function readTemplate(tplpath) {
    return fs.readFileSync(path.join(__dirname, "../../../" + tplpath), "utf8");
}
var ArenaTokenTemplates = /** @class */ (function () {
    function ArenaTokenTemplates(fungibleTokenAddress, arenaTokenAddress) {
        this.fungibleTokenAddress = fungibleTokenAddress;
        this.arenaTokenAddress = arenaTokenAddress;
    }
    ArenaTokenTemplates.prototype.sendArena = function (recipient, amount) {
        var template = readTemplate("transactions/arenaToken/send_arena.cdc");
        var code = resolveImports(template, new Map([
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
        var template = readTemplate("transactions/arenaToken/setup_account.cdc");
        var code = resolveImports(template, new Map([
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
        var template = readTemplate("scripts/arenaToken/get_balance.cdc");
        var code = resolveImports(template, new Map([
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