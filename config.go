package main

type Config struct {
	Title     string
	Subtitle  string
	Licence   string
	Port      int
	QuadStore quadStore
	UI        userInterface
	Vocab     vocabulary
}

type quadStore struct {
	Endpoint      string
	ShowAllGraphs bool
	Graphs        []string
	OpenTimeout   int
	ReadTimeout   int
}

type userInterface struct {
	ShowImages      bool
	NumImages       int
	ImagePredicates []string
}

type vocabulary struct {
	Enabled bool
	Dict    [][]string
}
