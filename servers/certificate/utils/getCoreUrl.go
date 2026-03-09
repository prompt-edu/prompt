package utils

import (
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

func GetCoreUrl() string {
	return promptSDK.GetEnv("SERVER_CORE_HOST", "http://localhost:8080")
}
