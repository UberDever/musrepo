package lib

import (
	"fmt"
	"net/url"
	"path"
)

type downloader_cmd struct {
	id  int
	out string
	cmd []string
}

func (c downloader_cmd) Dump() string {
	return fmt.Sprintf("%s '%s'", YT_DOWNLOADER, c.out)
}

func (c downloader_cmd) TrackId() int  { return c.id }
func (c downloader_cmd) In() string    { return "" }
func (c downloader_cmd) Out() string   { return c.out }
func (c downloader_cmd) Cmd() []string { return c.cmd }

func (c *MusRepo) downloader_command(track_id int, out string, url string) downloader_cmd {
	return downloader_cmd{
		id:  track_id,
		out: out + OUT_EXT,
		cmd: []string{
			YT_DOWNLOADER,
			"-x",
			"--no-progress",
			"-o", out,
			url,
		},
	}
}

func (c *MusRepo) Download(out_path string) ([]Command, error) {
	commands := make([]Command, 0, 8)
	for _, track := range c.music.Tracks {
		link, err := url.Parse(track.Url)
		if err != nil {
			return nil, err
		}
		if link.Scheme == "file" {
			continue
		}

		title := PathFriendly(track.Title)
		out := path.Join(out_path, title)
		commands = append(commands, c.downloader_command(track.id, out, link.String()))
	}
	return commands, nil
}
