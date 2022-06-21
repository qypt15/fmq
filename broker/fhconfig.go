package broker

type FhConfig struct {
	Worker   int       `json:"workerNum"`
	HTTPPort string    `json:"httpPort"`
	Host     string    `json:"host"`
	Port     string    `json:"port"`
	Cluster  RouteInfo `json:"cluster"`
	Router   string    `json:"router"`
	TlsHost  string    `json:"tlsHost"`
	TlsPort  string    `json:"tlsPort"`
	WsPath   string    `json:"wsPath"`
	WsPort   string    `json:"wsPort"`
	WsTLS    bool      `json:"wsTLS"`
	TlsInfo  TLSInfo   `json:"tlsInfo"`
	Debug    bool      `json:"debug"`
	Plugin   Plugins   `json:"plugins"`
}

func Test(args []string) (*FhConfig, error) {
	config := &FhConfig{}
	return config, nil
}
