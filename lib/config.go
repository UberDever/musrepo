package lib

const (
	CONVERTER     = "ffmpeg"
	YT_DOWNLOADER = "./yt-dlp"
	USAGE         = `
    Tool that uses provided list of tracks (currently in YAML format) and does the following:

    - Downloads the track from url (if present) or grabs from filesystem
    - Splits the track into pieces provided by Timestamps (if present) and converts
    it to (currently) opus
    - Optionally transfers tracks to provided location with repo integrity checking

    So, basically, I want to gradually increase number of tracks in the Music.yaml and
    gradually convert and transfer them
    `
	OUT_EXT = ".opus"
)
