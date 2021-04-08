package config

type Client struct {
	Connect struct {
		Scheme string
		Host   string
		Port   int
		Path   string
	}
}
