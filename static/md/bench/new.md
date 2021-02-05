---
Title: New Bench
Template: index.tmpl
---
# New Bench! What are you going to build?

## Let's add some details and we'll get you all set up :3

{{if .Error}}
<p>Error: {{.Error}}</p>
{{end}}

<form action="/bench/new" id="benchform" method="post">
    <label for="name">Bench Name</label><br>
    <input type="text" id="name" name="name" size="50" value=""><br><br>
    <label for="public">Show to other users?</label><br>
    <input type="checkbox" id="public" name="public"><br><br>
    <label for="description">Description</label><br>
    <textarea name="description" form="benchform" rows="5" cols="50" placeholder="(Optional text describing your awesome bench here)"></textarea><br><br>
    <input type="submit" value="Submit">
</form> 
