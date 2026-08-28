# Hoppr PowerShell Integration
# Add to $PROFILE: . /path/to/hop.ps1
#
# The binary is called "hop" directly. This wrapper adds:
#   1. Directory jumping: `hop <project>` does `cd` into the project path
#   2. Tab completion for project and command names

function hop {
    $allArgs = $args

    if ($allArgs.Count -eq 0) {
        & hop.exe
        return
    }

    $cmd = $allArgs[0]
    $nonJumpCmds = @(
        "add", "remove", "rm", "del", "delete",
        "list", "ls", "l",
        "create", "new", "drop", "rename", "mv",
        "setdefault", "default", "use",
        "import", "scan", "load",
        "doctor", "check", "diag",
        "update", "upgrade", "self-update",
        "help", "--help", "-h",
        "--version", "-v", "version"
    )

    # Single-argument invocation that is NOT a known command → treat as project jump
    if ($allArgs.Count -eq 1 -and $cmd -notin $nonJumpCmds) {
        $target = & hop.exe _get_path $cmd 2>$null
        if ($target -and (Test-Path $target)) {
            Set-Location $target
            return
        }
    }

    # Otherwise, pass everything through to the hop binary
    & hop.exe @allArgs
}

# PowerShell Tab Completion
Register-ArgumentCompleter -Native -CommandName hop -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $suggestions = & hop.exe _complete $wordToComplete 2>$null
    $suggestions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
