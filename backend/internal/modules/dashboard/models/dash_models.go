package models

type DashboardModule struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Icon     string `json:"icon"` // pode ser o nome de uma classe CSS ou uma URL
	Endpoint string `json:"endpoint"`
}
