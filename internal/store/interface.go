package store

type StoreInterface interface {
	Get(key string) string
	Set(key string, value string)
	Delete(key string)
	Exists(key string) bool
	Keys() []string
	Values() []string
	Size() int
	Clear()
	Copy() StoreInterface
	Merge(StoreInterface)
}
