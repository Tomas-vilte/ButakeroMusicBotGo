package common

type (
	// Event representa la estructura del evento recibido de GitHub
	Event struct {
		Action       string      `json:"action"`
		Release      Release     `json:"release,omitempty"`
		WorkFlowJobs WorkFlowJob `json:"workflow_job,omitempty"`
	}

	// ReleaseEvent representa la estructura del evento de lanzamiento recibido de GitHub
	ReleaseEvent struct {
		Action  string  `json:"action"`
		Release Release `json:"release"`
	}

	// WorkflowEvent representa la estructura del evento de flujo de trabajo recibido de GitHub
	WorkflowEvent struct {
		Action       string      `json:"action"`
		WorkFlowJobs WorkFlowJob `json:"workflow_job"`
	}

	// Release representa la estructura de una versión lanzada en GitHub
	Release struct {
		TagName   string `json:"tag_name"`
		Body      string `json:"body"`
		HtmlURL   string `json:"html_url"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}

	// WorkFlowJob representa la estructura de un trabajo de flujo de trabajo en GitHub
	WorkFlowJob struct {
		WorkFlowName string `json:"workflow_name"`
		ID           int    `json:"id"`
		HtmlURL      string `json:"html_url"`
		Status       string `json:"status"`
		Conclusion   string `json:"conclusion"`
		CreatedAt    string `json:"created_at"`
		StartedAt    string `json:"started_at"`
		CompletedAt  string `json:"completed_at"`
		Steps        []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"steps"`
	}
)
