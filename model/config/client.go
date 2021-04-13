package config

type Client struct {
	Connect struct {
		Scheme string `yaml:"scheme" long:"connect.scheme" description:"Scheme for connect" default:"ws"`
		Host   string `yaml:"host" long:"connect.host" description:"Server host" default:"localhost"`
		Port   int    `yaml:"port" long:"connect.port" description:"Server port" default:"8080"`
		Path   string `yaml:"path" long:"connect.path" description:"Listen path" default:"/task-manager-server"`
	} `yaml:"connect"`
}

func (Client) Pathfile() string {
	return "client.yaml"
}
