package completion

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GenerateBashCompletion 生成 Bash 补全脚本
func (cm *CompletionManager) GenerateBashCompletion(w io.Writer, cmdName string) error {
	script := fmt.Sprintf(`# bash completion for %s
_%s_completion() {
    local cur prev words cword
    
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion || return
    else
        COMPREPLY=()
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    fi

    local args=()
    for ((i=1; i<COMP_CWORD; i++)); do
        args+=("${COMP_WORDS[i]}")
    done

    local completions
    completions=$(%s __complete "${args[@]}" "$cur" 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        while IFS= read -r line; do
            COMPREPLY+=("$line")
        done < <(compgen -W "$completions" -- "$cur")
    fi
}

complete -F _%s_completion %s
`, cmdName, cmdName, cmdName, cmdName, cmdName)

	_, err := w.Write([]byte(script))
	return err
}

// GenerateZshCompletion 生成 Zsh 补全脚本
func (cm *CompletionManager) GenerateZshCompletion(w io.Writer, cmdName string) error {
	script := fmt.Sprintf(`#compdef %s

_%s() {
    local -a completions
    local curcontext="$curcontext" state line
    typeset -A opt_args

    local -a args
    local cur="${words[CURRENT]}"
    
    if (( CURRENT > 2 )); then
        args=("${words[@]:1:$((CURRENT-2))}")
    fi

    local output
    output=$(%s __complete --format=zsh "${args[@]}" "$cur" 2>/dev/null)
    
    if [[ $? -eq 0 ]]; then
        completions=(${(f)output})
        _describe 'completions' completions
    fi
}

_%s "$@"
`, cmdName, cmdName, cmdName, cmdName)

	_, err := w.Write([]byte(script))
	return err
}

// GenerateFishCompletion 生成 Fish 补全脚本
func (cm *CompletionManager) GenerateFishCompletion(w io.Writer, cmdName string) error {
	script := fmt.Sprintf(`# fish completion for %s
function __%s_complete
    set -l cmd (commandline -opc)
    set -l cur (commandline -ct)
    set -e cmd[1]
    %s __complete --format=fish $cmd $cur 2>/dev/null
end

complete -c %s -f -a "(__%s_complete)"
`, cmdName, cmdName, cmdName, cmdName, cmdName)

	_, err := w.Write([]byte(script))
	return err
}

// GenerateCompletion 生成补全脚本
func (cm *CompletionManager) GenerateCompletion(shell string, cname ...string) error {
	var cmdName string
	if len(cname) > 0 {
		cmdName = cname[0]
	} else {
		path, _ := os.Executable()
		_, cmdName = filepath.Split(path)
	}

	switch shell {
	case "bash":
		return cm.GenerateBashCompletion(os.Stdout, cmdName)
	case "zsh":
		return cm.GenerateZshCompletion(os.Stdout, cmdName)
	case "fish":
		return cm.GenerateFishCompletion(os.Stdout, cmdName)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}
