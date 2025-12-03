package config

type GlobalConfig struct {
	AppName       string `yaml:"AppName"`
	MODE          string `yaml:"Mode"` // dev或prod
	VERSION       string `yaml:"Version"`
	Host          string `yaml:"Host"`
	HTTPPort          string `yaml:"HttpPort"`
	Model struct {
		Model		 string `yaml:"Model"`
		Apikey        string `yaml:"Apikey"`
	} `yaml:"Model"`
	Milvus struct {
		Host     string `yaml:"Host"`
		Port     string `yaml:"Port"`
		CollectionName string `yaml:"CollectionName"`
	} `yaml:"Milvus"`
	Embedding struct {
		Model  string `yaml:"Model"`
		Apikey string `yaml:"Apikey"`
		Dimensions *int    `yaml:"Dimensions"`
	} `yaml:"Embedding"`
}
