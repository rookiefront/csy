package store

func Get[T any](s Store, key string) T {
	return _get[T](s.Db, s.Bucket, key)
}
