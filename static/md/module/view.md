---
Title: Module Name
Template: index.tmpl
---
# {{.Module.Name}} by {{.Author}}

{{.Module.Description}}

| [Repository]({{.Project.RepoURL}}) | [Add to Bench](/bench/add/{{.Module.UUID}}) |

<div style="border: lightgray 0.1em; border-radius: 0.5em; border-style: solid; padding: 0 0.5em;">
{{.Readme}}
</div>

{{if .Error}}
Error:
{{.Error}}
{{end}}
