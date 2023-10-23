package lib

type music struct {
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

func downloader_cmd(out_path string, url string) []string {
	return []string{
		YT_DOWNLOADER,
		"-x",
		"-o", out_path,
		url,
	}
}

func convert_cmd(in string, from string, to string, out string) []string {
	return []string{
		FFMPEG,
		"-i", in,
		"-ss", from,
		"-to", to,
		"-c", "copy",
		out,
	}
}
