package version

import "fmt"

var (
	AppVersion = "0.1.0"
	appName    = "Chapar"
)

func GetAppVersion() string {
	return AppVersion
}

func GetAgentName() string {
	return fmt.Sprintf("%s/%s", appName, AppVersion)
}
