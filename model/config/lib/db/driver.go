package db

type Driver struct {
	Db struct {
		Type string
		Name string
	}
	Connect struct {
		Host string
		Port int
	}
	Auth struct {
		User string
		Pass string
	}
	Extra map[string]interface{}
}
