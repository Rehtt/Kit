// Copyright (c) 2025 Rehtt
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

// bash模板
const bashTemplate = `# bash completion for {{.CommandName}}
# 生成时间: 自动生成
# 支持多级子命令补全

_{{.CommandName}}_completion() {
    local cur prev words cword
    _init_completion || return

    # 构建当前命令路径
    local cmd_path=()
    local i
    for ((i=1; i<cword; i++)); do
        # 跳过以 - 开头的flag
        if [[ "${words[i]}" != -* ]]; then
            cmd_path+=("${words[i]}")
        fi
    done

    # 根据命令路径深度提供补全
    local path_str="${cmd_path[*]}"
    
    case "$path_str" in
{{- range .AllCommands}}
        "{{.FullPath}}")
            # 命令: {{.FullPath}}
{{- if .Flags}}
            if [[ $cur == -* ]]; then
                COMPREPLY=($(compgen -W "{{flagsToString .Flags}}" -- "$cur"))
                return
            fi
{{- end}}
{{- if .SubCommands}}
            COMPREPLY=($(compgen -W "{{commandsToString .SubCommands}}" -- "$cur"))
{{- else}}
            # 没有子命令
{{- end}}
            ;;
{{- end}}
        *)
            # 根命令补全
{{- if .RootFlags}}
            if [[ $cur == -* ]]; then
                COMPREPLY=($(compgen -W "{{flagsToString .RootFlags}}" -- "$cur"))
                return
            fi
{{- end}}
{{- if .FirstLevelCommands}}
            COMPREPLY=($(compgen -W "{{commandsToString .FirstLevelCommands}}" -- "$cur"))
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
# 支持多级子命令补全

{{- define "renderZshSubCommand" -}}
{{- if .SubCommands}}
                    _arguments -C \
{{- range .Flags}}
                        {{zshFlagSpec .}} \
{{- end}}
                        '1: :->{{.Name}}_subcmds' \
                        '*::arg:->{{.Name}}_args' && return
                    
                    case $state in
                        {{.Name}}_subcmds)
                            local subcmds=(
{{- range .SubCommands}}
{{- if not .Hidden}}
                                '{{.Name}}:{{if .Instruction}}{{.Instruction}}{{else}}{{.Name}}{{end}}'
{{- end}}
{{- end}}
                            )
                            _describe 'subcommands' subcmds
                            ;;
                        {{.Name}}_args)
                            case $words[1] in
{{- range .SubCommands}}
{{- if not .Hidden}}
                                {{.Name}})
{{template "renderZshSubCommand" .}}
                                    ;;
{{- end}}
{{- end}}
                            esac
                            ;;
                    esac
{{- else}}
{{- if .Flags}}
                    _arguments \
{{- range $i, $flag := .Flags}}
                        {{zshFlagSpec $flag}}{{if ne $i (sub (len $.Flags) 1)}} \{{end}}
{{- end}} && return
{{- end}}
{{- end}}
{{- end}}

_{{.CommandName}}() {
    local context state line
    typeset -A opt_args

    _arguments -C \
{{- range .RootFlags}}
        {{zshFlagSpec .}} \
{{- end}}
{{- if .FirstLevelCommands}}
        '1: :->cmds' \
        '*::arg:->args' && return

    case $state in
        cmds)
            local commands=(
{{- range .FirstLevelCommands}}
{{- if not .Hidden}}
                '{{.Name}}:{{if .Instruction}}{{.Instruction}}{{else}}{{.Name}}{{end}}'
{{- end}}
{{- end}}
            )
            _describe 'commands' commands
            ;;
        args)
            case $words[1] in
{{- range .FirstLevelCommands}}
{{- if not .Hidden}}
                {{.Name}})
{{template "renderZshSubCommand" .}}
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
# 支持多级子命令补全

# 根命令的 flags
{{- range .RootFlags}}
{{fishFlagSpec $.CommandName . "__fish_use_subcommand"}}
{{- end}}

# 所有命令及其补全
{{- range $flatCmd := .AllCommands}}
{{- $pathLen := len $flatCmd.PathParts}}
{{- if eq $pathLen 1}}
{{- if not $flatCmd.Hidden}}
# 一级命令: {{$flatCmd.Name}}
complete -c {{$.CommandName}} -n '__fish_use_subcommand' -a {{$flatCmd.Name}} -d '{{if $flatCmd.Instruction}}{{$flatCmd.Instruction}}{{else}}{{$flatCmd.Name}}{{end}}'
{{- range $flatCmd.Flags}}
{{fishFlagSpec $.CommandName . (printf "__fish_seen_subcommand_from %s" $flatCmd.Name)}}
{{- end}}
{{- end}}
{{- else}}
{{- if not $flatCmd.Hidden}}
# {{$pathLen}}级命令: {{$flatCmd.FullPath}}
{{- $parentPath := slice $flatCmd.PathParts 0 (sub $pathLen 1)}}
complete -c {{$.CommandName}} -n '__fish_seen_subcommand_from {{index $parentPath (sub $pathLen 2)}}; and not __fish_seen_subcommand_from {{$flatCmd.Name}}' -a {{$flatCmd.Name}} -d '{{if $flatCmd.Instruction}}{{$flatCmd.Instruction}}{{else}}{{$flatCmd.Name}}{{end}}'
{{- range $flatCmd.Flags}}
{{fishFlagSpec $.CommandName . (printf "__fish_seen_subcommand_from %s" $flatCmd.Name)}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
`
