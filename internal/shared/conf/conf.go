package conf

import _ "embed"

//go:embed core_app.yaml
var CoreAppYaml []byte

//go:embed agent_app.yaml
var AgentAppYaml []byte
