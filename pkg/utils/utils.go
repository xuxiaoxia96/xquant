package utils

import (
	"github.com/bytedance/sonic"
	"os"
)

func GetEnv(key string) string {
	return GetEnvWithDefault(key, "")
}

func GetEnvWithDefault(key, defaultV string) string {
	if v, exist := os.LookupEnv(key); exist {
		return v
	} else {
		return defaultV
	}
}

func MustMarshalIndent(d interface{}) string {
	if d, err := sonic.ConfigDefault.MarshalIndent(d, "", "\t"); err != nil {
		return ""
	} else {
		return string(d)
	}
}
