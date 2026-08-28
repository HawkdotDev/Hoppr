# Hoppr Fish Integration
# Add to ~/.config/fish/config.fish: source /path/to/hop.fish

function hop
    set -l non_jump_cmds add remove rm del delete list ls l create new drop rename mv setdefault default use import scan load doctor check diag help --help -h --version -v version

    if test (count $argv) -eq 1
        if not contains -- $argv[1] $non_jump_cmds
            set -l target (command hoppr _get_path $argv[1] 2>/dev/null)
            if test -n "$target" -a -d "$target"
                cd "$target"
                return 0
            end
        end
    end

    command hoppr $argv
end

complete -c hop -f -a '(command hoppr _complete 2>/dev/null)'
complete -c hoppr -f -a '(command hoppr _complete 2>/dev/null)'
