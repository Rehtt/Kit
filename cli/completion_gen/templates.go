// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

// bash模板
const bashTemplate = `# bash completion for {{.CommandName}}
# 生成时间: 自动生成

_{{.CommandName}}_completion() {
    local cur prev words cword
    _init_completion || return

    local cmd="${words[0]}"
    local subcommand=""

    # 识别子命令
    local i
    for ((i=1; i<cword; i++)); do
        case "${words[i]}" in
{{- range .Commands}}
{{- if not .Hidden}}
            {{.Name}})
                subcommand="{{.Name}}"
                ;;
{{- range .SubCommands}}
{{- if not .Hidden}}
            {{.Name}})
                subcommand="{{.Name}}"
                ;;
{{- end}}
{{- end}}
{{- end}}
{{- end}}
        esac
    done

    # 补全逻辑
    case "$subcommand" in
{{- range .Commands}}
{{- if not .Hidden}}
        {{.Name}})
{{- if .Flags}}
            if [[ $cur == -* ]]; then
                COMPREPLY=($(compgen -W "{{flagsToString .Flags}}" -- "$cur"))
                return
            fi
{{- end}}
{{- if .SubCommands}}
            COMPREPLY=($(compgen -W "{{commandsToString .SubCommands}}" -- "$cur"))
{{- end}}
            ;;
{{- end}}
{{- end}}
        *)
            # 根命令补全
{{- if .RootFlags}}
            if [[ $cur == -* ]]; then
                COMPREPLY=($(compgen -W "{{flagsToString .RootFlags}}" -- "$cur"))
                return
            fi
{{- end}}
{{- if .Commands}}
            COMPREPLY=($(compgen -W "{{commandsToString .Commands}}" -- "$cur"))
{{- end}}
            ;;
    esac
}

complete -F _{{.CommandName}}_completion {{.CommandName}}
`

// zsh模板
const zshTemplate = `#compdef {{.CommandName}}
# zsh completion for {{.CommandName}}
# 生成时间: 自动生成

_{{.CommandName}}() {
    local context state line
    typeset -A opt_args

    _arguments -C \
{{- range .RootFlags}}
        {{zshFlagSpec .}} \
{{- end}}
{{- if .Commands}}
        '1: :->cmds' \
        '*::arg:->args'

    case $state in
        cmds)
            _values 'commands' \
{{- range $i, $cmd := .Commands}}
{{- if not $cmd.Hidden}}
{{- $desc := $cmd.Instruction}}
{{- if not $desc}}{{$desc = $cmd.Name}}{{end}}
                '{{$cmd.Name}}[{{$desc}}]'{{if ne $i (sub (len $.Commands) 1)}} \{{end}}
{{- end}}
{{- end}}
            ;;
        args)
            case $line[1] in
{{- range $cmdIdx, $cmd := .Commands}}
{{- if not $cmd.Hidden}}
                {{$cmd.Name}})
{{- if or $cmd.Flags $cmd.SubCommands}}
                    _arguments \
{{- range $cmd.Flags}}
                        {{zshFlagSpec .}} \
{{- end}}
{{- if $cmd.SubCommands}}
                        '1: :->subcmds'
                    case $state in
                        subcmds)
                            _values 'subcommands' \
{{- range $i, $sub := $cmd.SubCommands}}
{{- if not $sub.Hidden}}
{{- $subDesc := $sub.Instruction}}
{{- if not $subDesc}}{{$subDesc = $sub.Name}}{{end}}
                                '{{$sub.Name}}[{{$subDesc}}]'{{if ne $i (sub (len $cmd.SubCommands) 1)}} \{{end}}
{{- end}}
{{- end}}
                            ;;
                    esac
{{- end}}
{{- end}}
                    ;;
{{- end}}
{{- end}}
            esac
            ;;
    esac
{{- end}}
}

_{{.CommandName}} "$@"
`

// fish模板
const fishTemplate = `# fish completion for {{.CommandName}}
# 生成时间: 自动生成

{{- range .RootFlags}}
{{fishFlagSpec $.CommandName . ""}}
{{- end}}
{{- range $cmd := .Commands}}
{{- if not $cmd.Hidden}}
{{- $desc := $cmd.Instruction}}
{{- if not $desc}}{{$desc = $cmd.Name}}{{end}}
complete -c {{$.CommandName}} -n '__fish_use_subcommand' -a {{$cmd.Name}} -d '{{$desc}}'
{{- range $cmd.Flags}}
{{fishFlagSpec $.CommandName . (printf "__fish_seen_subcommand_from %s" $cmd.Name)}}
{{- end}}
{{- range $sub := $cmd.SubCommands}}
{{- if not $sub.Hidden}}
{{- $subDesc := $sub.Instruction}}
{{- if not $subDesc}}{{$subDesc = $sub.Name}}{{end}}
complete -c {{$.CommandName}} -n '__fish_seen_subcommand_from {{$cmd.Name}}' -a {{$sub.Name}} -d '{{$subDesc}}'
{{- end}}
{{- end}}
{{- end}}
{{- end}}
`
