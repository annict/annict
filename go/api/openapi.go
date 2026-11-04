// Package apiはREST API V2のOpenAPI記述 (`openapi.yaml`) を埋め込む。
//
// oapi-codegenが生成コードに埋め込む記述 (`apigen.GetSpec()`) は、operationIdが
// Goの名前 (`GetOpenAPIDescription` など) に書き換わっている。
// 利用者へ配信する記述には、正本のYAMLをそのまま使う。
package api

import _ "embed"

// OpenAPIDescriptionはREST API V2のOpenAPI記述 (YAML)
//
//go:embed openapi.yaml
var OpenAPIDescription []byte
