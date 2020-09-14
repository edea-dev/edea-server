---
Title: New Project
Template: index.tmpl
---
# New Project! Exciting!

## Let's add some details and we'll get you all set up :3

<form action="/project/new" id="projectform" method="post">
    <label for="name">Project Name</label><br>
    <input type="text" id="name" name="name" size="50" value=""><br><br>
    <label for="url">Repository URL</label><br>
    <input type="url" id="url" name="repourl" placeholder="https://github.com/..." size="50" value=""><br><br>
    <label for="description">Description</label><br>
    <textarea name="description" form="projectform" rows="5" cols="50" placeholder="(Optional text describing your awesome project here)"></textarea><br><br>
    <input type="submit" value="Submit">
</form> 
