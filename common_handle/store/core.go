package store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func _decodeValue(value interface{}) []byte {
	marshal, err := json.Marshal(value)
	if err == nil {
		return marshal
	}
	// 使用类型断言将接口变量转换为 []byte 类型
	if bytes, ok := value.([]byte); ok {
		return bytes
	}
	return nil
}

func _encodeValue[T any](value []byte) T {
	var result T
	err := json.Unmarshal(value, &result)
	if err == nil {
		return result
	}
	return result
}
func _get[T any | string](db *bolt.DB, bucket string, key string) T {
	var result T

	if db == nil {
		return result
	}
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return fmt.Errorf("桶 mybucket 不存在")
		}
		valueCopy := bucket.Get([]byte(key))
		result = _encodeValue[T](valueCopy)
		return nil
	})
	return result
}

func _set(db *bolt.DB, bucket string, key string, value interface{}) error {
	err := db.Update(func(tx *bolt.Tx) error {
		// 打开或创建一个名为 "mybucket" 的桶（Bucket）
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		bytes := _decodeValue(value)
		err = bucket.Put([]byte(key), bytes)
		return err
	})
	return err
}
func _createBucketIfNotExists(db *bolt.DB, bucket string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
	return err
}

func _getAll[T any](db *bolt.DB, bucket string) []T {
	var res []T
	_createBucketIfNotExists(db, bucket)
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		err := bucket.ForEach(func(k, v []byte) error {
			value := _encodeValue[T](v)
			res = append(res, value)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func _getAllKey(db *bolt.DB, bucketStr string) []string {
	var res []string
	_createBucketIfNotExists(db, bucketStr)
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketStr))
		err := bucket.ForEach(func(k, v []byte) error {
			res = append(res, string(k))
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return res
}

func _remove(db *bolt.DB, bucket string, key string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("桶 mybucket 不存在")
		}
		return b.Delete([]byte(key))
	})
}
