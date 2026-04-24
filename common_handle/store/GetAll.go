package store

func GetAll[T any](s Store) []T {
	return _getAll[T](s.Db, s.Bucket)
}

func GetAllKey(s Store) []string {

	return _getAllKey(s.Db, s.Bucket)
}
