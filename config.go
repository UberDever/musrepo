package musrepo

const (
	FFMPEG        = "ffmpeg"
	YT_DOWNLOADER = "yt-dlp"
	USAGE         = `
        Tool that uses provided list of tracks (currently in YAML format) and does the following:

        - Downloads the track from url (if present) or grabs from filesystem
        - Splits the track into pieces provided by Timestamps (if present) and converts
        it to (currently) opus
        - Optionally transfers tracks to provided location with repo integrity checking

        So, basically, I want to gradually increase number of tracks in the Music.yaml and
        gradually convert and transfer them
    `
)

func downloader_cmd(out_path string, url string) []string {
	return []string{
		YT_DOWNLOADER,
		"-x",
		"-o", out_path,
		url,
	}
}

func ffmpeg_cmd(in string, from string, to string, out string) []string {
	return []string{
		FFMPEG,
		"-i", in,
		"-ss", from,
		"-to", to,
		"-c", "copy",
		out,
	}
}
