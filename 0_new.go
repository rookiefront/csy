package csy

import (
	"github.com/front-ck996/csy/common_handle/cmd"
	"github.com/front-ck996/csy/common_handle/cron"
	"github.com/front-ck996/csy/common_handle/file"
)

var NewFile = file.NewFile

var NewCMD = cmd.NewCMD

func NewCron[T any]() *cron.Cron[T] {
	return cron.NewCron[T]()
}
