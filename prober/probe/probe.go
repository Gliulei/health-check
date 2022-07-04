package probe

import "time"

// Result 用于处理/返回探测结果
type Result string

const (
	Success Result = "success"
	Warning Result = "warning"
	Failure Result = "failure"
	Unknown Result = "unknown"
)

type Probe interface {
	Ping(url string, timeout time.Duration) error
}
