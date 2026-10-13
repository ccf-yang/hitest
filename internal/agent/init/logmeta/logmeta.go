package logmeta

import _ "embed"

// XLogJson contains operation-log metadata (interface name/format mapping).
//
//go:embed x-log.json
var XLogJson []byte
