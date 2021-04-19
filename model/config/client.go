package config

type Client struct {
	Connect struct {
		Scheme string `yaml:"scheme" env:"CONNECT_SCHEME" long:"connect.scheme" description:"Scheme for connect" default:"ws"`
		Host   string `yaml:"host" env:"CONNECT_HOST" long:"connect.host" description:"Server host" default:"localhost"`
		Port   int    `yaml:"port" env:"CONNECT_PORT" long:"connect.port" description:"Server port" default:"8080"`
		Path   string `yaml:"path" env:"CONNECT_PATH" long:"connect.path" description:"Listen path" default:"/task-manager-server"`
	} `yaml:"connect"`
}

func (Client) Pathfile() string {
	return "client.yaml"
}
