---
Title: Profile
Template: index.tmpl
---
# Welcome {{ .User.Profile.DisplayName }}

Hello World!

<form action="/profile" method="post">
<input type="hidden" id="id" name="id" value="{{ .User.Profile.ID }}">
<label for="display_name">Display Name:</label><br>
<input type="text" id="display_name" name="display_name" value="{{ .User.Profile.DisplayName }}"><br>
<label for="location">Location:</label><br>
<input type="text" id="location" name="location" value="{{ .User.Profile.Location }}"><br>
<label for="biography">Biography:</label><br>
<input type="text" id="biography" name="biography" value="{{ .User.Profile.Biography }}"><br>
<label for="avatar">Avatar:</label><br>
<input type="text" id="avatar" name="avatar" value="{{ .User.Profile.Avatar }}"><br><br>
<input type="submit" value="Submit">
</form>
