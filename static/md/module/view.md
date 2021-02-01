---
Title: Module Name
Template: index.tmpl
---
# {{.Module.Name}} by {{.Author}}

{{if .Error}}<p>Error: {{.Error}}</p>{{end}}
{{.Module.Description}}

| <a href="{{.Module.RepoURL}}">Repository</a> | <a href="/bench/add/{{.Module.ID}}">Add to Bench</a> |

<div style="border: lightgray 0.1em; border-radius: 0.5em; border-style: solid; padding: 0 0.5em;">
{{.Readme}}
</div>
