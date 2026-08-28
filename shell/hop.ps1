# Hoppr PowerShell Integration
# Add to $PROFILE: . /path/to/hop.ps1

function hop {
    # Collect all arguments as an array (avoids named-param / $args conflict)
    $allArgs = $args

    if ($allArgs.Count -eq 0) {
        & hoppr
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
        "help", "--help", "-h",
        "--version", "-v", "version"
    )

    # Single-argument invocation that is NOT a known command → treat as project jump
    if ($allArgs.Count -eq 1 -and $cmd -notin $nonJumpCmds) {
        $target = & hoppr _get_path $cmd 2>$null
        if ($target -and (Test-Path $target)) {
            Set-Location $target
            return
        }
    }

    # Otherwise, pass everything through to the hoppr binary
    & hoppr @allArgs
}

# PowerShell Tab Completion
Register-ArgumentCompleter -Native -CommandName hop,hoppr -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $suggestions = & hoppr _complete $wordToComplete 2>$null
    $suggestions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
