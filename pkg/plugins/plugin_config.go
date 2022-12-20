package plugins

// Config
//
//	plugins:
//	  label:
//	  - --known-values=aaa,bbb,ccc
type Configuration struct {
	Owner   string                `json:"owner"`
	Repo    string                `json:"repo"`
	Plugins []PluginConfiguration `json:"plugins"`
}

type PluginConfiguration struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}
