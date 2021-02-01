---
Title: Bench Name
Template: bench_view.tmpl
---
# {{.Bench.Name}} by {{.Author}}

{{.Bench.Description}}

{{if .Error}}
<p>Error: {{.Error}}</p>
{{end}}
