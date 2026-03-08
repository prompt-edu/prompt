package utils

import (
	promptSDK "github.com/ls1intum/prompt-sdk"
)

func GetCoreUrl() string {
	return promptSDK.GetEnv("SERVER_CORE_HOST", "http://localhost:8080")
}
