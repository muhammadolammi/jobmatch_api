package auth

import "os"

func getSecureMode() bool {
	enviroment := os.Getenv("ENV")
	securemode := false
	if enviroment == "deployment" {
		securemode = true
	}
	return securemode
}
