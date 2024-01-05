package store

// Store is a map of strings to strings.
// This means that when you assign a map to a new variable, both variables refer to the same underlying data.
type Store map[string]string

func NewStore() Store {
	return make(Store)
}

func (s Store) Get(key string) string {
	return s[key]
}

func (db Store) Set(key string, value string) {
	db[key] = value
}

func (db Store) Delete(key string) {
	delete(db, key)
}

func (db Store) Exists(key string) bool {
	_, ok := db[key]
	return ok
}

func (db Store) Keys() []string {
	keys := make([]string, len(db))
	i := 0
	for key := range db {
		keys[i] = key
		i++
	}
	return keys
}

func (db Store) Values() []string {
	values := make([]string, len(db))
	i := 0
	for _, value := range db {
		values[i] = value
		i++
	}
	return values
}

func (db Store) Size() int {
	return len(db)
}

func (db Store) Clear() {
	for key := range db {
		delete(db, key)
	}
}

func (db Store) Copy() Store {
	newDB := make(Store)
	for key, value := range db {
		newDB[key] = value
	}
	return newDB
}

func (db Store) Merge(newDB Store) {
	for key, value := range newDB {
		db[key] = value
	}
}

func (db Store) String() string {
	str := "{\n"
	for key, value := range db {
		str += "\t" + key + ": " + value + "\n"
	}
	str += "}"
	return str
}

func (db Store) Print() {
	println(db.String())
}
