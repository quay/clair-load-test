package utils

// Type to store the test config.
type TestConfig struct {
	Containers []string      `json:"containers"`
	Concurrency  int     `json:"concurrency"`
	ESHost		string		 `json:"eshost"`
	ESPort		string 		 `json:"esport"`
	ESIndex		string 		 `json:"esindex"`
	Host       string        `json:"host"`
	HitSize		int 	`json:"hitsize"`
	IndexDelete     bool     `json:"delete"`
	Psk        string        `json:"-"`
	Uuid 	   string 		 `json:"uuid"`
}