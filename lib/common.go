package lib

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Music struct {
	Tracks []track `yaml:"Music"`
}

type track struct {
	id         int
	Type       string `yaml:"Type"`
	Title      string `yaml:"Title"`
	Url        string `yaml:"Url"`
	End        string `yaml:"End"`
	Timestamps string `yaml:"Timestamps"`
}

type MusRepo struct {
	music *Music
}

func NewMusRepo(m *Music) MusRepo {
	return MusRepo{
		music: m,
	}
}

type Command interface {
	TrackId() int
	Cmd() []string
}

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

	for i := range music.Tracks {
		music.Tracks[i].id = i
	}
	return &music, nil
}

func FormatCommand(c Command) string {
	return fmt.Sprintf("%d: ", c.TrackId()) + "'" + strings.Join(c.Cmd(), "', '") + "'"
}
