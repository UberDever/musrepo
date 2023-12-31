package lib

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"regexp"
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
	Dump() string
	In() string
	Out() string
	Cmd() []string
}

func ExecCommand(c Command) (error, []byte) {
	command := c.Cmd()
	cmd := exec.Command(command[0], command[1:]...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return err, nil
	}
	std_out, _ := io.ReadAll(stdout)
	err_out, _ := io.ReadAll(stderr)
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("failed: %s\n%s", c.Dump(), err_out), nil
	}
	return nil, std_out
}

func FormatCommand(c Command) string {
	return fmt.Sprintf("%d: ", c.TrackId()) + "'" + strings.Join(c.Cmd(), "', '") + "'"
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
func SetupLogHandlers(verbose bool) {
	ReplaceAttr := func(group []string, a slog.Attr) slog.Attr {
		if a.Key == "time" || a.Key == "level" {
			return slog.Attr{}
		}
		return slog.Attr{Key: a.Key, Value: a.Value}
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: ReplaceAttr})
	slog.SetDefault(slog.New(handler))

	if verbose {
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: ReplaceAttr})
		slog.SetDefault(slog.New(handler))
	}
}

func CreateOutDirs(cmd Command) {
	dirname := path.Dir(cmd.Out())
	if _, err := os.Stat(dirname); errors.Is(err, os.ErrNotExist) {
		slog.Info(fmt.Sprintf("Creating directory '%s'", dirname))
		os.MkdirAll(dirname, os.ModePerm)
	}

}

func PathFriendly(str string) string {
	to_remove := []rune(`/:*?"<>|`)
	sb := strings.Builder{}
	for _, ch := range str {
		found := false
		for i := range to_remove {
			if ch == to_remove[i] {
				found = true
				break
			}
		}
		if !found {
			sb.WriteRune(ch)
		}
	}
	re := regexp.MustCompile(`\b\w+\.[^\.]`)
	res := re.ReplaceAllString(sb.String(), "")
	if res == "" {
		panic(fmt.Sprintf("Path is empty! %s", str))
	}
	return res
}
