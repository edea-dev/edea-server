---
Title: Profile
Template: index.tmpl
---
# Welcome {{ .Profile.DisplayName }}

Hello World!

<form action="/profile" method="post">
<input type="hidden" id="id" name="id" value="{{ .Profile.ID }}">
<label for="display_name">Display Name:</label><br>
<input type="text" id="display_name" name="display_name" value="{{ .Profile.DisplayName }}"><br>
<label for="location">Location:</label><br>
<input type="text" id="location" name="location" value="{{ .Profile.Location }}"><br>
<label for="biography">Biography:</label><br>
<input type="text" id="biography" name="biography" value="{{ .Profile.Biography }}"><br>
<label for="avatar">Avatar:</label><br>
<input type="text" id="avatar" name="avatar" value="{{ .Profile.Avatar }}"><br><br>
<input type="submit" value="Submit">
</form>
