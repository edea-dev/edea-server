---
Title: Project Name
Template: index.tmpl
---
# {{.Project.Name}} by {{.Author}}

{{.Project.Description}}

| [Repository]({{.Project.RepoURL}}) | [Add to Bench](/bench/add/{{.Project.UUID}}) |

<div style="border: lightgray 0.1em; border-radius: 0.5em; border-style: solid; padding: 0 0.5em;">
{{.Readme}}
</div>

{{if .Error}}
Error:
{{.Error}}
{{end}}
