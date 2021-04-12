package config

type Server struct {
	Listen struct {
		Ip   string `yaml:"host" env:"LISTEN_IP" long:"listen.ip"   description:"Listen ip"`
		Port string `yaml:"port" env:"LISTEN_PORT" long:"listen.port"   description:"Listen port"`
		Path string `yaml:"path" env:"LISTEN_PATH" long:"listen.path" description:"Listen path"`
	} `yaml:"listen"`
}

func (Server) Pathfile() string {
	return "server.yaml"
}
