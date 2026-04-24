package store

func Set(s Store, key string, value interface{}) error {
	return _set(s.Db, s.Bucket, key, value)
}
