package config

func Init() {

}

func GetString(name string) string {
	if name == "port" {
		return "0.0.0.0:8080"
	}
	return ""
}
