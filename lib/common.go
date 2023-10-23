package lib

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Music struct {
	Tracks []track `yaml:"Music"`
}

type track struct {
	Type       string `yaml:"Type"`
	Title      string `yaml:"Title"`
	Url        string `yaml:"Url"`
	End        string `yaml:"End"`
	Timestamps string `yaml:"Timestamps"`
}

type command []string

func LoadMusic(path string) (*Music, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var music Music
	err = yaml.Unmarshal(contents, &music)
	if err != nil {
		return nil, err
	}
	return &music, nil
}
