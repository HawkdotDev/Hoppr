# Hoppr PowerShell Integration
# Add to $PROFILE: . /path/to/hop.ps1

function hop {
    param([string]$cmd, [string]$arg1, [string]$arg2)
    $nonJumpCmds = @("add", "remove", "rm", "del", "delete", "list", "ls", "l", "create", "new", "drop", "rename", "mv", "setdefault", "default", "use", "import", "scan", "load", "doctor", "check", "diag", "help", "--help", "-h", "--version", "-v", "version")

    if ($args.Count -eq 1 -and $cmd -notin $nonJumpCmds) {
        $target = & hoppr _get_path $cmd 2>$null
        if ($target -and (Test-Path $target)) {
            Set-Location $target
            return
        }
    }

    & hoppr @args
}

# PowerShell Tab Completion
Register-ArgumentCompleter -Native -CommandName hop,hoppr -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $suggestions = & hoppr _complete $wordToComplete 2>$null
    $suggestions | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}
