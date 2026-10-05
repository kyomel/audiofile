package models

type Audio struct {
	Id       string   `json:"id"`
	Path     string   `json:"path"`
	Metadata Metadata `json:"metadata"`
	Status   string   `json:"status"`
	Error    []string `json:"error,omitempty"`
}

type Metadata struct {
	Tags       []Tag  `json:"tags"`
	Transcript string `json:"transcript"`
	Duration   int    `json:"duration"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
