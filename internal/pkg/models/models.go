package models

type DBConf struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
}

type AccountingServiceConf struct {
	Scheme          string `json:"scheme" yaml:"scheme"`
	Host            string `json:"host" yaml:"host"`
	Port            string `json:"port" yaml:"port"`
	HealthCheckPath string `json:"health_check_path" yaml:"health_check_path"`
}
