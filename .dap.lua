vim.print("wtf")

---@class dap.Configuration[]
---@field type string
---@field request "launch"|"attach"
---@field name string
---@field [string] any
return {
    {
        type = "go",
        request = "launch",
        name = "debug ./cmd/mihomo-update/",
        program = "${workspaceFolder}/cmd/mihomo-update",
        args = function()
            local args = vim.fn.input("Args (enter to ommit): ")
            vim.print(args)
            vim.print(require("utils").shell_split(args))
            return require("utils").shell_split(args)
        end,
        cwd = "${workspaceFolder}",
    }
}
