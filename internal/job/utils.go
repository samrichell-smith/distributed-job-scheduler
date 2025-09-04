package job

import "sync"

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

var resultMutex sync.Mutex

func (j *Job) addPartialSum(partial int) {
	resultMutex.Lock()
	defer resultMutex.Unlock()

	var res LargeArraySumResult
	if j.Result != nil {
		res = j.Result.(LargeArraySumResult)
	}
	res.Sum += partial
	j.Result = res
}
