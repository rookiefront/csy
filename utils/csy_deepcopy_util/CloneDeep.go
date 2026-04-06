package csy_deepcopy_util

import (
	"bytes"
	"encoding/gob"

	"github.com/front-ck996/csy/0_extend/deepcopy"
)

func CloneDeep[T any](dst T, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func CloneDeepMohae[T any](src interface{}) T {
	newVal := deepcopy.Copy(src)
	return newVal.(T)
}
