package reports

// Type to store the test config.
type TestConfig struct {
	Containers []string      `json:"containers"`
	Psk        string        `json:"-"`
	Host       string        `json:"host"`
	IndexDelete     bool     `json:"delete"`
	HitSize		float64 	`json:"hitsize"`
	Concurrency  float64     `json:"concurrency"`
	ESHost		string		 `json:"eshost"`
	ESPort		string 		 `json:"esport"`
	ESIndex		string 		 `json:"esindex"`
}