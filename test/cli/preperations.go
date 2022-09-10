package test

import (
	"os"
)

func SetEnvironmentVariables() {
	os.Setenv("PM_API_URL", "https://192.168.67.50:8006/api2/json")
	os.Setenv("PM_USER", "root@pam")
	os.Setenv("PM_PASS", "Enter123!")
}
