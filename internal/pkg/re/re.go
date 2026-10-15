// Package re: 移植自 legacy shared/re（只留磁盘解析用到的）
package re

import (
	"fmt"
	"regexp"
	"sync"
)

const DiskKeyValuePattern = `([A-Za-z0-9_]+)=("([^"\\]|\\.)*"|[^ \t]+)`

var (
	regexMap = map[string]*regexp.Regexp{}
	mu       sync.Mutex
)

// GetRegex: 带缓存的正则编译
func GetRegex(pattern string) *regexp.Regexp {
	mu.Lock()
	defer mu.Unlock()
	regex, exists := regexMap[pattern]
	if !exists {
		var err error
		regex, err = regexp.Compile(pattern)
		if err != nil {
			panic(fmt.Sprintf("invalid regex pattern: %s", pattern))
		}
		regexMap[pattern] = regex
	}
	return regex
}
