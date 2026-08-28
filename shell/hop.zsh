# Hoppr Zsh Integration
# Add to ~/.zshrc: source /path/to/hop.zsh

hop() {
    local -a non_jump_cmds
    non_jump_cmds=(add remove rm del delete list ls l create new drop rename mv setdefault default use import scan load doctor check diag update upgrade self-update help --help -h --version -v version)

    if [ "$#" -eq 1 ]; then
        if [[ ! " ${non_jump_cmds[*]} " =~ " $1 " ]]; then
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

# Dynamic Zsh Completion
_hop_zsh_completions() {
    local -a suggestions
    suggestions=(${(f)"$(command hop _complete "$words[CURRENT]" 2>/dev/null)"})
    _describe 'projects and lists' suggestions
}

compdef _hop_zsh_completions hop
