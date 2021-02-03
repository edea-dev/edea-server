package middleware

var devErrorTmpl = `
  <div class="container content">
    <div class="row">
      <div class="col-md-12">
        <h3>Error Trace</h3>
        <pre>{{.error}}</pre>
        <h3>Stack Trace</h3>
        <pre>{{.stacktrace}}</pre>
        <h3>Context</h3>
        <pre>{{printf "%+v" .context}}</pre>
        <h3>Parameters</h3>
        <pre>{{printf "%+v" .vars}}</pre>
        <h3>Headers</h3>
        <pre>{{printf "%+v" .headers}}</pre>
        <h3>Form</h3>
        <pre>{{printf "%+v" .form}}</pre>
        <h3>Routes</h3>
        <table class="table table-striped">
          <thead>
            <tr text-align="left">
              <th class="centered">METHOD</th>
              <th>PATH</th>
              <th>NAME</th>
              <th>HANDLER</th>
            </tr>
          </thead>
          <tbody>
            {{range .routes}}
              <tr>
                <td class="centered">
                  {{ .Methods }}
                </td>
                <td>
                    <a href="{{ .Path }}">{{ .Path }}</a>
                </td>
                <td>
                  {{ .Path }}
                </td>
                <td><code> {{ .Func }} </code></td>
              </tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </div>
    <div class="foot"> <span> Powered by ADHD and <a href="https://en.wikipedia.org/wiki/List_of_methylphenidate_analogues">Methylphenidate</a></span></div>
  </div>
`
var prodErrorTmpl = `
<!DOCTYPE html>
<html>
<head>
<style>h1,p.powered{text-align:center}body{background:#ECECEC;padding-top:25px;font-family:helvetica neue,helvetica,sans-serif;color:#333}.card{box-sizing:border-box;width:440px;min-width:270px;margin:0 auto;padding:10px 25px 35px 10px;background:#FFF;box-shadow:0 2px 4px 0 rgba(185,185,185,.28);border-radius:5px}.card p{max-width:320px;margin:15px auto}h1{font-size:22px}hr{border:.5px solid #D72727;width:180px}p.powered{font-family:HelveticaNeue-Light;font-size:12px;color:#333}@media (max-width:600px){.card{width:100%;display:block}}</style>
</head>
<body>
<div class="container">
	<div class="card">
		<h1>We're Sorry!</h1>
		<hr>
		<p>It looks like something went wrong! Don't worry, we are aware of the problem and are looking into it.</p>
		<p>Sorry if this has caused you any problems. Please check back again later.</p>
	</div>
	<p class="powered">powered by <a href="https://gobuffalo.io">gobuffalo.io</a></p>
</div>
</body>
</html>
`

var prodNotFoundTmpl = `
<!DOCTYPE html>
<html>
<head>
<style>h1,p.powered{text-align:center}body{background:#ECECEC;padding-top:25px;font-family:helvetica neue,helvetica,sans-serif;color:#333}.card{box-sizing:border-box;width:440px;min-width:270px;margin:0 auto;padding:10px 25px 35px 10px;background:#FFF;box-shadow:0 2px 4px 0 rgba(185,185,185,.28);border-radius:5px}.card p{max-width:320px;margin:15px auto}h1{font-size:22px}hr{border:.5px solid #1272E2;width:180px}p.powered{font-family:HelveticaNeue-Light;font-size:12px;color:#333}@media (max-width:600px){.card{width:100%;display:block}}</style>
</head>
<body>
<div class="container">
	<div class="card">
		<h1>Not Found</h1>
		<hr>
		<p>The page you're looking for does not exist, you may have mistyped the address or the page may have been moved.</p>
	</div>
	<p class="powered">powered by <a href="https://gobuffalo.io">gobuffalo.io</a></p>
</div>
</body>
</html>
`
