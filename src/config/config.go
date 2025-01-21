package config

type Config struct {
	HTTP      HTTP `json:"http"`
	Databases struct {
		Printful Database `json:"printful"`
		Images   Database `json:"images"`
	} `json:"databases"`
	Printful Printful `json:"printful"`
}

type HTTP struct {
	Port          int    `json:"port"`
	HttpsKeyFile  string `json:"https_key_file"`
	HttpsCertFile string `json:"https_cert_file"`
}

type Database struct {
	ConnectURI string `json:"connect_uri"`
	DBName     string `json:"db_name"`
	BucketName string `json:"bucket_name"`
}

type Printful struct {
	AccessToken     string `json:"access_token"`
	SimulateMockup  bool   `json:"simulate_mockup"`
	SimulateTaskKey string `json:"simulate_task_key"`
	TaskInterval    int    `json:"task_interval"`
	MockupDirectory string `json:"mockup_directory"`
	ImagesURL       string `json:"images_url"`
}
