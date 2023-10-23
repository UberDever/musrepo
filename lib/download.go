package lib

import "net/url"

func downloader_cmd(out_path string, url string) []string {
	return []string{
		YT_DOWNLOADER,
		"-x",
		"-o", out_path,
		url,
	}
}

func Download(m Music, out_path string) ([]command, error) {
	commands := make([]command, 0, 8)
	for _, track := range m.Tracks {
		link, err := url.Parse(track.Url)
		if err != nil {
			return nil, err
		}
		if link.Scheme == "file" {
			continue
		}

		commands = append(commands, downloader_cmd(out_path, link.String()))
	}
	return commands, nil
}
