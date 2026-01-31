package csy

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"sync"
	"time"
)

type Cron[T any] struct {
	C       *cron.Cron
	TaskMap map[cron.EntryID]string
	MetaMap map[cron.EntryID]T
	mu      sync.RWMutex // 读写锁

}
type CronEntries struct {
	cron.Entry
	Expression string
}

// NewCron 创建计划任务
func NewCron[T any]() *Cron[T] {
	c := Cron[T]{
		TaskMap: map[cron.EntryID]string{},
	}
	// 日志输出
	logger := log.New(log.Writer(), "[CRON] ", log.LstdFlags)
	// 创建时区（确保与服务器 Linux 系统时间对齐）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	c.C = cron.New(
		cron.WithLocation(loc),
		cron.WithChain(
			cron.Recover(cron.PrintfLogger(logger)),
		),
		// 开启秒支持
		cron.WithSeconds(),
	)

	//// 示例 A: 每 5 秒执行一次 (类似 */5 * * * *)
	//_, _ = c.AddFunc("*/5 * * * * *", func() {
	//	fmt.Printf("[%s] 任务启动：执行清理缓存...\n", time.Now().Format("15:04:05"))
	//})

	c.C.Start()

	return &c
}

func (c *Cron[T]) ParseTimeToExpression(t time.Time) string {
	// 注意：Cron 的周是从 0 (周日) 到 6 (周六)
	// t.Weekday() 在 Go 中默认也是 0-6，所以可以直接转换
	return fmt.Sprintf("%d %d %d %d %d %d",
		t.Second(),
		t.Minute(),
		t.Hour(),
		t.Day(),
		int(t.Month()),
		int(t.Weekday()),
	)
}

// AddTimeTask 添加时间任务
func (c *Cron[T]) AddTimeTask(t time.Time, f func(), meta T) error {
	expression := c.ParseTimeToExpression(t)
	// 这里 meta 可以用于日志记录或者预处理，但不存入 TaskMap
	return c.AddExpressionTask(expression, f, meta)
}

// AddExpressionTask 添加时间任务
func (c *Cron[T]) AddExpressionTask(expression string, f func(), meta T) error {
	id, err := c.C.AddFunc(expression, f)
	if err == nil {
		c.mu.Lock()
		c.TaskMap[id] = expression
		c.MetaMap[id] = meta
		c.mu.Unlock()
	}
	return err
}

func (c *Cron[T]) GetTaskList() []CronEntries {

	entries := c.C.Entries()
	list := make([]CronEntries, 0, len(entries))

	for _, entry := range entries {
		exp := c.TaskMap[entry.ID]
		list = append(list, CronEntries{
			Entry:      entry,
			Expression: exp,
		})
	}
	return list
}

func (c *Cron[T]) RemoveTask(id cron.EntryID) {
	c.C.Remove(id)
	c.mu.Lock()
	delete(c.TaskMap, id)
	c.mu.Unlock()
}
