package config

type Server struct {
	Listen struct {
		Ip   string `yaml:"host" env:"LISTEN_IP" long:"listen.ip"   description:"Listen ip" default:"0.0.0.0"`
		Port string `yaml:"port" env:"LISTEN_PORT" long:"listen.port"   description:"Listen port" default:"8080"`
		Path string `yaml:"path" env:"LISTEN_PATH" long:"listen.path" description:"Listen path" default:"/task-manager-server"`
	} `yaml:"listen"`
}

func (Server) Pathfile() string {
	return "server.yaml"
}
