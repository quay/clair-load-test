package manifests

// Type used to fetch clair manifest hash.
type ManifestHash struct {
    ManifestHash string `json:"hash"`
}

// Type used to store results of channel.
type result struct {
    index int
	container string
    err   error
}