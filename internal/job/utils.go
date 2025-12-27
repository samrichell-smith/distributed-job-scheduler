package job

import (
	"fmt"
	"time"
)

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (j *Job) addPartialSum(partial int) {
	j.resultMu.Lock()
	defer j.resultMu.Unlock()

	var res LargeArraySumResult
	if j.Result != nil {
		// try value type
		if r, ok := j.Result.(LargeArraySumResult); ok {
			res = r
		} else if rp, ok := j.Result.(*LargeArraySumResult); ok {
			res = *rp
		} else {
			// incompatible type — reset to zeroed result
			res = LargeArraySumResult{}
		}
	}
	res.Sum += partial
	j.Result = res
}

func ResizeImage(url string, width, height int) string {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("%s_resized_%dx%d", url, width, height)
}
