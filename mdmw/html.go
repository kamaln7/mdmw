package mdmw

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
