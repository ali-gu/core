package tool

type ListToolsResult struct {
	ID          string
	Type        string
	DisplayName string
	Config      map[string]any
}
