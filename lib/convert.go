package lib

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

type convert_cmd struct {
	id  int
	in  string
	out string
	cmd []string
}

func (c convert_cmd) Dump() string {
	return fmt.Sprintf("%s '%s' -> '%s'", CONVERTER, c.in, c.out)
}

func (c convert_cmd) TrackId() int  { return c.id }
func (c convert_cmd) In() string    { return c.in }
func (c convert_cmd) Out() string   { return c.out }
func (c convert_cmd) Cmd() []string { return c.cmd }

func (c *MusRepo) convert_command(track_id int, in string, from string, to string, out string) convert_cmd {
	return convert_cmd{
		id:  track_id,
		in:  in,
		out: out,
		cmd: []string{
			CONVERTER,
			"-hide_banner",
			"-loglevel", "error",
			"-i", in,
			"-ss", from,
			"-to", to,
			"-c", "copy",
			out,
		},
	}
}

func (c *MusRepo) Convert(in_dir string, out_dir string) ([]convert_cmd, error) {
	commands := make([]convert_cmd, 0, 8)
	for _, track := range c.music.Tracks {
		track_parts, err := convert_timestamps(track.Timestamps, track.End)
		if err != nil {
			return nil, err
		}

		for _, part := range track_parts {
			title := PathFriendly(track.Title)
			in := path.Join(in_dir, title) + OUT_EXT
			out := path.Join(out_dir, title, PathFriendly(part.name)) + OUT_EXT
			cmd := c.convert_command(track.id, in, part.start, part.end, out)
			commands = append(commands, cmd)
		}
	}
	return commands, nil
}

type track_part struct {
	name  string
	start string
	end   string
}

func convert_timestamps(timestamps string, end string) ([]track_part, error) {
	time_regex := regexp.MustCompile(`^(?:[0-9]+:)?(?:[0-5]?[0-9]):(?:[0-5]?[0-9])$`)
	looks_like_time := func(s string) bool {
		return time_regex.MatchString(s)
	}

	names := []string{}
	starts := []string{}
	ends := []string{}

	lines := strings.Split(strings.TrimSpace(timestamps), "\n")
	for index, line := range lines {
		parts := strings.Fields(strings.TrimSpace(line))
		time_part := -1
		for i, part := range parts {
			if looks_like_time(part) {
				time_part = i
				break
			}
		}
		if time_part == -1 {
			return nil, fmt.Errorf("no time found in track part: %s", line)
		}

		time := parts[time_part]
		parts = append(parts[:time_part], parts[time_part+1:]...)
		entry := fmt.Sprintf("%03d %s", index, strings.Join(parts, " "))
		names = append(names, entry)
		starts = append(starts, time)
	}
	ends = append(ends, starts[1:]...)
	ends = append(ends, end)

	if len(names) != len(starts) ||
		len(names) != len(ends) ||
		len(starts) != len(ends) {
		return nil, fmt.Errorf("lengths must be equal %v\n%v\n%v", names, starts, ends)
	}
	track_parts := []track_part{}
	for i := range names {
		track_parts = append(track_parts, track_part{
			name:  names[i],
			start: starts[i],
			end:   ends[i],
		})
	}

	return track_parts, nil
}
