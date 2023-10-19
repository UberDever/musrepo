package musrepo

type Music struct {
	Tracks []Track `yaml:"Music"`
}

type Track struct {
	Type       string `yaml:"Type"`
	Title      string `yaml:"Title"`
	Url        string `yaml:"Url"`
	End        string `yaml:"End"`
	Timestamps string `yaml:"Timestamps"`
}
