package lib

type Storage struct {
	Db struct {
		DriverId string `yaml:"driver_id" env:"DB_DRIVER_ID" long:"db.driver_id" description:"Db driver id" default:"storage"`
	}
}

func (Storage) Pathfile() string {
	return "storage.yaml"
}
