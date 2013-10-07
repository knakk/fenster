package main

type Config struct {
	License    string
	LicenseURL string
	BaseURI    string
	Port       int
	QuadStore  quadStore
	UI         userInterface
	Vocab      vocabulary
}

type quadStore struct {
	Endpoint     string
	OpenTimeout  int
	ReadTimeout  int
	ResultsLimit int
}

type userInterface struct {
	ShowImages      bool
	NumImages       int
	ImagePredicates []string
	TitlePredicates []string
	RootRedirectTo  string
}

type vocabulary struct {
	Enabled bool
	Dict    [][]string
}
