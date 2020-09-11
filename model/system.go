package model

const (
	SYSTEM_INSTALLATION_DATE_KEY = "InstallationDate"
)

type System struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
