package common

type (
	// Event representa la estructura del evento recibido de GitHub
	Event struct {
		Action  string  `json:"action"`
		Release Release `json:"release"`
	}

	// Release representa la estructura de una versión lanzada en GitHub
	Release struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
	}
)
