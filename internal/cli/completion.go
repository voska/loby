package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/voska/loby/internal/errfmt"
)

// CompletionCmd emits a shell completion script. Sourced into the user's shell,
// it turns `loby <TAB>` into a working completion experience. The script is
// generated from the current Kong model so command additions stay in sync.
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell to generate completions for." enum:"bash,zsh,fish,powershell"`
}

// Run renders the requested shell script.
func (c *CompletionCmd) Run(g *Globals, k *kong.Kong) error {
	cmds, flags := walkSchema(k.Model.Node)
	switch c.Shell {
	case "bash":
		return writeString(g, bashCompletion(cmds, flags))
	case "zsh":
		return writeString(g, zshCompletion(cmds, flags))
	case "fish":
		return writeString(g, fishCompletion(cmds))
	case "powershell":
		return writeString(g, powershellCompletion(cmds))
	default:
		return errfmt.Wrap(errfmt.UsageError, errors.New("unsupported shell: "+c.Shell))
	}
}

// walkSchema collects top-level command names and global flag names from the
// Kong model. Subcommands are picked up dynamically via the script's runtime
// call to `loby schema --json`, so the generated file is small.
func walkSchema(n *kong.Node) (commands, flags []string) {
	for _, child := range n.Children {
		if child.Hidden {
			continue
		}
		commands = append(commands, child.Name)
	}
	for _, f := range n.Flags {
		if f.Hidden {
			continue
		}
		flags = append(flags, "--"+f.Name)
	}
	return commands, flags
}

func writeString(g *Globals, s string) error {
	_, err := fmt.Fprint(g.Stdout, s)
	if err != nil {
		return fmt.Errorf("write completion: %w", err)
	}
	return nil
}

func bashCompletion(commands, flags []string) string {
	return fmt.Sprintf(`# loby bash completion. Source with:
#   loby completion bash > /etc/bash_completion.d/loby   # system-wide
#   loby completion bash > ~/.loby-completion.bash && echo 'source ~/.loby-completion.bash' >> ~/.bashrc

_loby() {
    local cur prev words cword
    _init_completion || return

    local top_commands="%s"
    local global_flags="%s"

    if [[ ${cur} == --* ]]; then
        COMPREPLY=( $(compgen -W "${global_flags}" -- "${cur}") )
        return 0
    fi

    if [[ ${cword} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${top_commands}" -- "${cur}") )
        return 0
    fi

    # Subcommand completion: ask the binary at runtime.
    local subcommands
    subcommands=$(loby schema "${words[1]}" --json 2>/dev/null | jq -r '.subcommands[]?.name' 2>/dev/null)
    if [[ -n "$subcommands" ]]; then
        COMPREPLY=( $(compgen -W "${subcommands}" -- "${cur}") )
    fi
}

complete -F _loby loby
`, strings.Join(commands, " "), strings.Join(flags, " "))
}

func zshCompletion(commands, flags []string) string {
	return fmt.Sprintf(`#compdef loby
# loby zsh completion. Source with:
#   loby completion zsh > "${fpath[1]}/_loby"

_loby() {
    local context state state_descr line
    typeset -A opt_args

    local -a top_commands global_flags
    top_commands=(%s)
    global_flags=(%s)

    _arguments -C \
        '1: :->cmd' \
        '*::arg:->args'

    case $state in
        cmd)
            _describe 'command' top_commands
            ;;
        args)
            local subs
            subs=( ${(f)"$(loby schema $words[1] --json 2>/dev/null | jq -r '.subcommands[]?.name' 2>/dev/null)"} )
            if (( ${#subs[@]} )); then
                _describe 'subcommand' subs
            else
                _values 'flag' $global_flags
            fi
            ;;
    esac
}

_loby "$@"
`, strings.Join(quoteList(commands), " "), strings.Join(quoteList(flags), " "))
}

func fishCompletion(commands []string) string {
	var b strings.Builder
	b.WriteString("# loby fish completion. Source with:\n")
	b.WriteString("#   loby completion fish > ~/.config/fish/completions/loby.fish\n\n")
	for _, c := range commands {
		fmt.Fprintf(&b, "complete -c loby -n '__fish_use_subcommand' -a '%s'\n", c)
	}
	b.WriteString(`
function __loby_subcommands
    set -l cmd (commandline -opc)
    if test (count $cmd) -ge 2
        loby schema $cmd[2] --json 2>/dev/null | jq -r '.subcommands[]?.name' 2>/dev/null
    end
end

complete -c loby -n '__fish_seen_subcommand_from ' -a '(__loby_subcommands)'
`)
	return b.String()
}

func powershellCompletion(commands []string) string {
	return fmt.Sprintf(`# loby powershell completion. Add to your $PROFILE:
#   loby completion powershell | Out-String | Invoke-Expression

Register-ArgumentCompleter -Native -CommandName 'loby' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @(%s)
    $tokens = $commandAst.CommandElements
    if ($tokens.Count -le 2) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } |
            ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
    } else {
        $sub = & loby schema $tokens[1].Value --json 2>$null | ConvertFrom-Json
        if ($sub.subcommands) {
            $sub.subcommands.name | Where-Object { $_ -like "$wordToComplete*" } |
                ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
        }
    }
}
`, strings.Join(quoteList(commands), ", "))
}

func quoteList(xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = "'" + x + "'"
	}
	return out
}
