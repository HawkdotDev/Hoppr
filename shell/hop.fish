# Hoppr Fish Integration
# Add to ~/.config/fish/config.fish: source /path/to/hop.fish

function hop
    set -l non_jump_cmds add remove rm del delete list ls l create new drop rename mv setdefault default use import scan load doctor check diag help --help -h --version -v version

    if test (count $argv) -eq 1
        if not contains -- $argv[1] $non_jump_cmds
            set -l target (command hop _get_path $argv[1] 2>/dev/null)
            if test -n "$target" -a -d "$target"
                cd "$target"
                return 0
            end
        end
    end

    command hop $argv
end

complete -c hop -f -a '(command hop _complete 2>/dev/null)'
