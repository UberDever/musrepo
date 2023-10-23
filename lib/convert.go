package lib

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

func Convert(m music, in_dir string, out_dir string) ([]command, error) {
	commands := make([]command, 0, 8)
	for _, track := range m.Tracks {
		track_parts, err := convert_timestamps(track.Timestamps, track.End)
		if err != nil {
			return nil, err
		}

		for _, part := range track_parts {
			in := path.Join(in_dir, track.Title) + OUT_EXT
			out := path.Join(out_dir, part.name) + OUT_EXT
			cmd := convert_cmd(in, part.start, part.end, out)
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

	forbiddens := []string{"."}
	names := []string{}
	starts := []string{}
	ends := []string{}

	lines := strings.Split(strings.TrimSpace(timestamps), " ")
	for index, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), " ")
		var time_part *string = nil
		for _, part := range parts {
			if looks_like_time(part) {
				*time_part = part
				break
			}
		}
		if time_part == nil {
			return nil, fmt.Errorf("no time found in track part: %s", line)
		}

		forbiddens = append(forbiddens, *time_part)
		legal_parts := make([]string, 0, len(parts))
		for _, forbidden := range forbiddens {
			for _, part := range parts {
				if !strings.Contains(part, forbidden) {
					legal_parts = append(legal_parts, part)
				}
			}
		}
		if len(legal_parts) == 0 {
			return nil, fmt.Errorf("no legal parts are found in parts: %v", parts)
		}

		n := fmt.Sprintf("%3d", index)
		names = append(names, n+" "+strings.Join(parts, " "))
		starts = append(starts, *time_part)
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
