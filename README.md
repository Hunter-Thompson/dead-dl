# Dead-DL

A command-line tool to download music from [relisten.org](https://relisten.net), which streams live music recordings from the Internet Archive's Live Music Archive.

## Features

- Download shows by band and year
- Support for multiple audio formats (MP3, FLAC)
- Automatically organizes downloads by band/year/show
- Handles multiple sources per show
- Skips already downloaded files

## Known Issues
- A lot of FLAC downloads fail due to archive.org restrictions (401/403); MP3 is more reliable, download all mp3 first then try FLAC if desired
- Some shows may not have the requested format available (FLAC mostly)

## Installation

```bash
go install github.com/Hunter-Thomson/dead-dl@latest
```

## Usage

### Basic Usage

Download all shows for a band in a specific year:

```bash
./dead-dl -band grateful-dead -year 1965
```

### Options

- `-band`: Band slug (e.g., `grateful-dead`, `phish`, `moe`). Default: `grateful-dead`
- `-year`: Year to download (required)
- `-output`: Output directory for downloads. Default: `./downloads`
- `-format`: Preferred format: `flac`, `mp3`, or `both`. Default: `mp3`
- `-highest-rated`: Whether to select the highest rated source for each show. Default: `false`

### Examples

Download Grateful Dead shows from 1977 in MP3 format:

```bash
./dead-dl -band grateful-dead -year 1977 -format mp3
```

Download Phish shows from 1995 in FLAC format:

```bash
./dead-dl -band phish -year 1995 -format flac -output ~/music/phish
```

Download all formats:

```bash
./dead-dl -band grateful-dead -year 1965 -format both
```

Run:

```
./dead-dl -band grateful-dead -year 1967 -output ./test-downloads --highest-rated
Found 23 shows for grateful-dead in 1967

[1/23] Processing show: 1967-01-14 at Polo Field, Golden Gate Park, San Francisco, CA, USA
  Selected highest rated source with avg rating 7.72
  Source [1/1]: archive.org identifier: gd67-01-14.sbd.vernon.9108.sbeok.shnf
    - No FLAC files found, falling back to MP3...
    - Downloading Morning Dew.mp3...
      Morning Dew.mp3 100% |████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████| (12/12 MB, 1.1 MB/s)
    - ✓ Downloaded Morning Dew.mp3
    - Downloading Viola Lee Blues.mp3...
      Viola Lee Blues.mp3  41% |████████████████████████████████████████████████████████████████████████                                                                                                        | (6.4/15 MB, 1.0 MB/s) [6s:8s]^
```

## How It Works

1. Fetches show listings from the relisten.org API for the specified band and year
2. For each show, retrieves source information (which includes archive.org identifiers)
3. Downloads audio files directly from archive.org in the requested format
4. Organizes files in the directory structure: `{output}/{band}/{year}/{show-date}/`

## Notes

- The tool respects rate limits by adding small delays between downloads
- Files that already exist are skipped (useful for resuming interrupted downloads)
- Some shows may have multiple sources (different recordings); each source is saved in a separate directory

## License

This project is provided as-is for personal use. Please respect the terms of use of both relisten.org and archive.org.

