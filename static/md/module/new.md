---
Title: New Module
Template: index.tmpl
---
# New Module! Exciting!

## Let's add some details and we'll get you all set up :3

<form action="/module/new" id="moduleform" method="post">
    <label for="name">Module Name</label><br>
    <input type="text" id="name" name="name" size="50" value=""><br><br>
    <label for="url">Repository URL</label><br>
    <input type="url" id="url" name="repourl" placeholder="https://github.com/..." size="50" value=""><br><br>
    <label for="description">Description</label><br>
    <textarea name="description" form="moduleform" rows="5" cols="50" placeholder="(Optional text describing your awesome module here)"></textarea><br><br>
    <input type="submit" value="Submit">
</form> 
