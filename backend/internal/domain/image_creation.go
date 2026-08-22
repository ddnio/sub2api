package domain

type ImageCreationTemplateDocument struct {
	SchemaVersion int                                     `json:"schema_version"`
	Title         string                                  `json:"title"`
	Summary       string                                  `json:"summary"`
	Category      string                                  `json:"category"`
	Tags          []string                                `json:"tags"`
	Prompt        string                                  `json:"prompt"`
	InputMode     string                                  `json:"input_mode"`
	CoverAlt      string                                  `json:"cover_alt"`
	Defaults      ImageCreationTemplateDefaults           `json:"defaults"`
	Source        *ImageCreationTemplateSourceAttribution `json:"source,omitempty"`
}

type ImageCreationTemplateDefaults struct {
	Size         string `json:"size"`
	Quality      string `json:"quality"`
	OutputFormat string `json:"output_format"`
}

type ImageCreationTemplateSourceAttribution struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	License string `json:"license,omitempty"`
	Notes   string `json:"notes,omitempty"`
}
