package store

import (
	"fmt"
	"sync"

	"github.com/front-ck996/csy"
	bolt "go.etcd.io/bbolt"
)

type Store struct {
	DbName string
	Db     *bolt.DB
	Init   StoreInit
	Bucket string
}
type StoreInit struct {
	DbDir      string
	DbName     string
	DbFullFile string
}

var dbs = map[string]Store{}
var lock sync.Mutex

func init() {
	dbs = map[string]Store{}
}

// NewStore 获取存储对象
func NewStore(init StoreInit) (Store, error) {
	if _, ok := dbs[init.DbName]; !ok {
		if init.DbDir == "" {
			init.DbDir = "_dbs/"
		}
		_s := Store{
			Init:   init,
			DbName: init.DbName,
			Bucket: "default",
		}
		lock.Lock()
		init.DbFullFile = fmt.Sprintf("%s/%s", init.DbDir, init.DbName)
		if err := csy.NewFile().FileExistsCreateDir(init.DbFullFile); err != nil {
			return Store{}, err
		}
		db, err := bolt.Open(init.DbDir+init.DbName, 0666, nil)
		if err != nil {
			return Store{}, err
		}
		_s.Db = db
		dbs[init.DbName] = _s
		lock.Unlock()
	}
	return dbs[init.DbName], nil
}
func (s *Store) ClearBucket() {
	s.Bucket = "default"
}
func (s *Store) SetBucket(bucket string) Store {
	cloneStore := Store{
		Bucket: bucket,
		Db:     s.Db,
		DbName: s.DbName,
		Init:   s.Init,
	}
	return cloneStore
}
