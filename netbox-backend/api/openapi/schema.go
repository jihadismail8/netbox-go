// Package openapi embeds the generated public REST contract.
package openapi

import _ "embed"

// Schema is generated from the validated capability profile.
//
//go:embed netbox-go-v1.yaml
var Schema []byte
