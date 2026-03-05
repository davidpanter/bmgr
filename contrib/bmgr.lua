-- bmgr.lua — Browse keybindings from Neovim
--
-- Install: copy to ~/.config/nvim/plugin/bmgr.lua
--
-- Usage:
--   :Bmgr          — browse all bindings
--   :Bmgr neovim   — browse bindings filtered to an app
--   <leader>k      — browse all bindings (keymap)
--
-- Requires: fzf, bmgr

local function bmgr(app)
  local cmd = "bmgr"
  if app and app ~= "" then
    cmd = cmd .. " --app " .. vim.fn.shellescape(app)
  end

  local buf = vim.api.nvim_create_buf(false, true)
  local width  = math.floor(vim.o.columns * 0.85)
  local height = math.floor(vim.o.lines   * 0.80)
  local win = vim.api.nvim_open_win(buf, true, {
    relative  = "editor",
    width     = width,
    height    = height,
    col       = math.floor((vim.o.columns - width)  / 2),
    row       = math.floor((vim.o.lines   - height) / 2),
    style     = "minimal",
    border    = "rounded",
    title     = " bmgr ",
    title_pos = "center",
  })
  vim.fn.termopen(cmd, {
    on_exit = function()
      if vim.api.nvim_win_is_valid(win) then
        vim.api.nvim_win_close(win, true)
      end
    end,
  })
  vim.cmd("startinsert")
end

vim.api.nvim_create_user_command("Bmgr", function(opts)
  bmgr(opts.args)
end, { nargs = "?" })

vim.keymap.set("n", "<leader>k", function() bmgr("neovim") end, { desc = "Browse Neovim keybindings (bmgr)" })
