package config

type key string

const (
	KeyUUID    = key("uuid")
	KeyMetrics = key("metrics")
	KeyLogger  = key("logger")
)
