package main

import (
	_ "embed"
	"net/http"
)


var openAPISpec []byte

func handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(openAPISpec)
}


func handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerUIPage))
}

const swaggerUIPage = `<!doctype html>
<html lang="fr">
<head>
	<meta charset="utf-8" />
	<title>BarterSwap API — Documentation</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
	<script>
		window.onload = () => {
			SwaggerUIBundle({
				url: "/openapi.json",
				dom_id: "#swagger-ui",
			});
		};
	</script>
</body>
</html>`
