package mdmw

import (
	"html/template"
)

const HTMLServerError = `
	<!DOCTYPE html>
	<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>500 internal server error</title>
		<style>
		p {
			text-align:center; font-family: -apple-system, "Helvetica Neue", "Lucida Grande", Helvetica, Arial, sans-serif; color: #666; font-size: 24px;
		}
		strong {
			color: #444;
		}
		</style>
	</head>
	<body>
		<p><strong>500</strong> internal server error</p>
	</body>
	</html>	
`
const HTMLNotFound = `
	<!DOCTYPE html>
	<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>404 not found</title>
		<style>
		p {
			text-align:center; font-family: -apple-system, "Helvetica Neue", "Lucida Grande", Helvetica, Arial, sans-serif; color: #666; font-size: 24px;
		}
		strong {
			color: #444;
		}
		</style>
	</head>
	<body>
		<p><strong>404</strong> not found</p>
	</body>
	</html>	
`
const HTMLForbidden = `
	<!DOCTYPE html>
	<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>403 forbidden</title>
		<style>
		p {
			text-align:center; font-family: -apple-system, "Helvetica Neue", "Lucida Grande", Helvetica, Arial, sans-serif; color: #666; font-size: 24px;
		}
		strong {
			color: #444;
		}
		</style>
	</head>
	<body>
		<p><strong>403</strong> forbidden</p>
	</body>
	</html>	
`

const HTMLOutput = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
		<style>/*! normalize.css v8.0.0 | MIT License | github.com/necolas/normalize.css */html{line-height:1.15;-webkit-text-size-adjust:100%}body{margin:0}h1{font-size:2em;margin:.67em 0}hr{box-sizing:content-box;height:0;overflow:visible}pre{font-family:monospace,monospace;font-size:1em}a{background-color:transparent}abbr[title]{border-bottom:0;text-decoration:underline;text-decoration:underline dotted}b,strong{font-weight:bolder}code,kbd,samp{font-family:monospace,monospace;font-size:1em}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative;vertical-align:baseline}sub{bottom:-0.25em}sup{top:-0.5em}img{border-style:none}button,input,optgroup,select,textarea{font-family:inherit;font-size:100%;line-height:1.15;margin:0}button,input{overflow:visible}button,select{text-transform:none}button,[type="button"],[type="reset"],[type="submit"]{-webkit-appearance:button}button::-moz-focus-inner,[type="button"]::-moz-focus-inner,[type="reset"]::-moz-focus-inner,[type="submit"]::-moz-focus-inner{border-style:none;padding:0}button:-moz-focusring,[type="button"]:-moz-focusring,[type="reset"]:-moz-focusring,[type="submit"]:-moz-focusring{outline:1px dotted ButtonText}fieldset{padding:.35em .75em .625em}legend{box-sizing:border-box;color:inherit;display:table;max-width:100%;padding:0;white-space:normal}progress{vertical-align:baseline}textarea{overflow:auto}[type="checkbox"],[type="radio"]{box-sizing:border-box;padding:0}[type="number"]::-webkit-inner-spin-button,[type="number"]::-webkit-outer-spin-button{height:auto}[type="search"]{-webkit-appearance:textfield;outline-offset:-2px}[type="search"]::-webkit-search-decoration{-webkit-appearance:none}::-webkit-file-upload-button{-webkit-appearance:button;font:inherit}details{display:block}summary{display:list-item}template{display:none}[hidden]{display:none}/*! minimis.css | MIT License | github.com/kamaln7/minimis.css */body{max-width:960px;margin:0 auto;padding:1rem .75rem;font-size:1.125rem;line-height:1.5;font-family:-apple-system,BlinkMacSystemFont,avenir next,avenir,segoe ui,helvetica neue,helvetica,ubuntu,roboto,noto,arial,sans-serif;text-rendering:optimizeLegibility;-webkit-font-smoothing:antialiased;color:#111}a{text-decoration:none}a:hover,a:focus{text-decoration:underline}a:visited{color:#00f}h1{margin-top:1.25rem;margin-bottom:1rem}h2{margin-top:1rem;margin-bottom:.875rem}h3{margin-top:.875rem;margin-bottom:.75rem}h4{margin-top:.875rem;margin-bottom:.75rem}h5{margin-top:.75rem;margin-bottom:.75rem}h6{margin-top:.75rem;margin-bottom:.75rem}pre{background:#f7f8fa;padding:.5rem}pre>code{white-space:pre-wrap}hr{border:0;height:1px;background-color:#c7c7c7}img{max-width:100%}</style>
	</head>
	<body>
{{.Body}}
	</body>
</html>
`

func (s *Server) SetOutputTemplate(source string) error {
	if source == "" {
		source = HTMLOutput
	}

	tmpl, err := template.New("").Parse(source)
	s.outputTmpl = tmpl
	return err
}
