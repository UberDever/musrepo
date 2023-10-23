package lib

import "net/url"

type downloader_cmd struct {
	id  int
	Out string
	cmd []string
}

func (c downloader_cmd) TrackId() int  { return c.id }
func (c downloader_cmd) Cmd() []string { return c.cmd }

func (c *MusRepo) downloader_command(track_id int, out_path string, url string) downloader_cmd {
	return downloader_cmd{
		id:  track_id,
		Out: out_path,
		cmd: []string{
			YT_DOWNLOADER,
			"-x",
			"-o", out_path,
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

		commands = append(commands, c.downloader_command(track.id, out_path, link.String()))
	}
	return commands, nil
}
