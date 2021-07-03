package middleware

// SPDX-License-Identifier: EUPL-1.2

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
