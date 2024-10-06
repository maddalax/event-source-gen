package main

import (
	"fmt"
	path2 "path"
)

func GetSourcingPath(path string) string {
	return path2.Join(Config.SourcingDir, path)
}

func PanicF(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func Unique[T any](slice []T, key func(item T) string) []T {
	var result []T
	seen := make(map[string]bool)
	for _, v := range slice {
		k := key(v)
		if _, ok := seen[k]; !ok {
			seen[k] = true
			result = append(result, v)
		}
	}
	return result
}
