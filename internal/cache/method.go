package cache

type ReportCache interface {
	Set(key string, value string) error
	Get(key string) error
}
