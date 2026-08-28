# Hoppr Bash Integration
# Add to ~/.bashrc: source /path/to/hop.bash

hop() {
    local non_jump_cmds=("add" "remove" "rm" "del" "delete" "list" "ls" "l" "create" "new" "drop" "rename" "mv" "setdefault" "default" "use" "import" "scan" "load" "doctor" "check" "diag" "help" "--help" "-h" "--version" "-v" "version")

    # If single argument provided and not a known subcommand, attempt project jump
    if [ "$#" -eq 1 ]; then
        local is_subcmd=0
        for cmd in "${non_jump_cmds[@]}"; do
            if [ "$1" = "$cmd" ]; then
                is_subcmd=1
                break
            fi
        done

        if [ "$is_subcmd" -eq 0 ]; then
            local target
            target=$(command hop _get_path "$1" 2>/dev/null)
            if [ -n "$target" ] && [ -d "$target" ]; then
                cd "$target" || return
                return 0
            fi
        fi
    fi

    command hop "$@"
}

# Dynamic Bash Completion
_hop_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local suggestions
    suggestions=$(command hop _complete "$cur" 2>/dev/null)
    COMPREPLY=( $(compgen -W "$suggestions" -- "$cur") )
}

complete -F _hop_completions hop
