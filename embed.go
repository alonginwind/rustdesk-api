package rustdeskapi

import _ "embed"

// DefaultConfigYAML is conf/config.yaml as shipped. config.Init uses it for keys missing from the admin's file.
//
//go:embed conf/config.yaml
var DefaultConfigYAML []byte
