package config

import (
	common "github.com/Arinashin3/LogTail_exporter"
)

type (
	Matcher interface {
		Match(common.logtailAttributes) bool
	}
)
