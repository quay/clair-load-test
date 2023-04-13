package commons

// Type to store the test config.
type TestConfig struct {
	Containers     []string `json:"containers"`
	Concurrency    int      `json:"concurrency"`
	TestRepoPrefix string   `json:"testrepoprefix"`
	ESHost         string   `json:"eshost"`
	ESPort         string   `json:"esport"`
	ESIndex        string   `json:"esindex"`
	Host           string   `json:"host"`
	HitSize        int      `json:"hitsize"`
	IndexDelete    bool     `json:"delete"`
	PSK            string   `json:"-"`
	UUID           string   `json:"uuid"`
}
