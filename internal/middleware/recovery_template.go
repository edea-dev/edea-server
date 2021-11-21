package middleware

// SPDX-License-Identifier: EUPL-1.2

var devErrorTmpl = `
<!DOCTYPE html>
<html class="h-100" lang="en">

<head>
  <title>EDeA - Error</title>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description"
    content="EDeA is a community design exchange platform for easily reusable, modular electronics development." />

  <link rel="icon" type="image/svg+xml" href="/icon.svg">

  <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
  <link rel="manifest" href="/site.webmanifest">
  <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#d72638">
  <meta name="apple-mobile-web-app-title" content="EDeA">
  <meta name="application-name" content="EDeA">
  <meta name="msapplication-TileColor" content="#da532c">
  <meta name="theme-color" content="#ffffff">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
  <link rel="stylesheet" href="/css/custom.css">
  <link rel="stylesheet" href="/css/bootstrap-table.css">

  <style>
    img {
      max-width: 100%;
    }
  </style>
</head>

<body class="d-flex flex-column h-100">
  <main role="main">
    <div class="container content">
      <div class="row">
        <div class="col-md-12">
          <h3>Error Trace</h3>
          <pre>{{.error}}</pre>
          <h3>Stack Trace</h3>
          <pre>{{.stacktrace}}</pre>
          <h3>Keys</h3>
          <pre>{{printf "%+v" .keys}}</pre>
          <h3>Parameters</h3>
          <pre>{{printf "%+v" .vars}}</pre>
          <h3>Headers</h3>
          <table class="table table-striped">
            <thead>
              <tr>
                <th>Header</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              {{range $key, $value := .headers}}
              <tr>
                <td>{{ $key }}</td>
                <td>{{ $value }}</td>
              </tr>
              {{end}}
            </tbody>
          </table>
          <h3>Form</h3>
          <table class="table table-striped">
            <thead>
              <tr>
                <th>Key</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              {{range $key, $value := .form}}
              <tr>
                <td>{{ $key }}</td>
                <td>{{ $value }}</td>
              </tr>
              {{end}}
            </tbody>
          </table>
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
      <div class="foot"> <span> Powered by ADHD and caffeine.</span></div>
    </div>
  </main>
</body>

</html>
`
