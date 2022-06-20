package dev

type Config struct {
	Worker   int    `json:"workerNum"`
	HTTPPort string `json:"httpPort"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Router   string `json:"router"`
	TlsHost  string `json:"tlsHost"`
	TlsPort  string `json:"tlsPort"`
	WsPath   string `json:"wsPath"`
	WsPort   string `json:"wsPort"`
	WsTLS    bool   `json:"wsTLS"`
	Debug    bool   `json:"debug"`
}

func RetTest(router string) (*Config, error) {
	config := &Config{}
	config.Router = router
	// config = "666"
	return config, nil
}
