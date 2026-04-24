package store

func Remove(s Store, key string) error {
	return _remove(s.Db, s.Bucket, key)
}
