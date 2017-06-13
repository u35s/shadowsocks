package zlib

import "strconv"

func FlowType(tp int) string {
	if tp == 0 {
		return "请求"
	} else if tp == 1 {
		return "回应"
	} else {
		return "未知"
	}
}

func Atoi(s string) int {
	if i, err := strconv.ParseUint(s, 10, 0); err == nil {
		return int(i)
	}
	return 0
}

func Atou(s string) uint {
	if i, err := strconv.ParseUint(s, 10, 0); err == nil {
		return uint(i)
	}
	return 0
}

func Atof(s string) float32 {
	if f, err := strconv.ParseFloat(s, 32); err == nil {
		return float32(f)
	}
	return 0
}
