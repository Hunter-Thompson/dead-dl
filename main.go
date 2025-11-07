package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Logger wraps multiple log.Logger instances for different log levels
type Logger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	file  *os.File
}

// NewLogger creates a new logger with time-based log file
func NewLogger() (*Logger, error) {
	// Create logs directory if it doesn't exist
	logsDir := "./logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Generate time-based log filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logsDir, fmt.Sprintf("dead-dl_%s.log", timestamp))

	// Create or open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	return &Logger{
		debug: log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds),
		info:  log.New(multiWriter, "[INFO]  ", log.Ldate|log.Ltime),
		warn:  log.New(multiWriter, "[WARN]  ", log.Ldate|log.Ltime),
		error: log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		file:  logFile,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Debug logs debug messages
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debug.Printf(format, v...)
}

// Info logs info messages
func (l *Logger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	l.warn.Printf(format, v...)
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	l.error.Printf(format, v...)
}

// Fatal logs error messages and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.error.Printf(format, v...)
	os.Exit(1)
}

// Printf logs a formatted message to both console and file
func (l *Logger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Print(msg)
	if l.file != nil {
		l.file.WriteString(msg)
	}
}

// Println logs a message with newline to both console and file
func (l *Logger) Println(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Println(msg)
	if l.file != nil {
		l.file.WriteString(msg + "\n")
	}
}

var logger *Logger

const (
	RelistenAPIBase = "https://api.relisten.net/api/v2"
	ArchiveAPIBase  = "https://archive.org"
)

type Show struct {
	Date        string   `json:"date"`
	DisplayDate string   `json:"display_date"`
	UUID        string   `json:"uuid"`
	Venue       Venue    `json:"venue"`
	Sources     []Source `json:"sources,omitempty"`
}

type Venue struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type ShowsResponse struct {
	Shows []Show `json:"shows"`
}

type Source struct {
	ReviewCount        int64       `json:"review_count"`
	Sets               []Set       `json:"sets"`
	Links              []Link      `json:"links"`
	ShowID             int64       `json:"show_id"`
	ShowUUID           string      `json:"show_uuid"`
	Show               interface{} `json:"show"`
	Description        string      `json:"description"`
	TaperNotes         string      `json:"taper_notes"`
	Source             string      `json:"source"`
	Taper              string      `json:"taper"`
	Transferrer        string      `json:"transferrer"`
	Lineage            string      `json:"lineage"`
	FLACType           string      `json:"flac_type"`
	ArtistID           int64       `json:"artist_id"`
	ArtistUUID         string      `json:"artist_uuid"`
	VenueID            int64       `json:"venue_id"`
	VenueUUID          string      `json:"venue_uuid"`
	Venue              interface{} `json:"venue"`
	DisplayDate        string      `json:"display_date"`
	IsSoundboard       bool        `json:"is_soundboard"`
	IsRemaster         bool        `json:"is_remaster"`
	HasJamcharts       bool        `json:"has_jamcharts"`
	AvgRating          float64     `json:"avg_rating"`
	NumReviews         int64       `json:"num_reviews"`
	NumRatings         interface{} `json:"num_ratings"`
	AvgRatingWeighted  float64     `json:"avg_rating_weighted"`
	Duration           float64     `json:"duration"`
	UpstreamIdentifier string      `json:"upstream_identifier"`
	UUID               string      `json:"uuid"`
	ID                 int64       `json:"id"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
}

type Set struct {
	SourceID   int64   `json:"source_id"`
	SourceUUID string  `json:"source_uuid"`
	ArtistID   int64   `json:"artist_id"`
	ArtistUUID string  `json:"artist_uuid"`
	UUID       string  `json:"uuid"`
	Index      int64   `json:"index"`
	IsEncore   bool    `json:"is_encore"`
	Name       string  `json:"name"`
	Tracks     []Track `json:"tracks"`
	ID         int64   `json:"id"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type Track struct {
	SourceID      int64       `json:"source_id"`
	SourceUUID    string      `json:"source_uuid"`
	SourceSetID   int64       `json:"source_set_id"`
	SourceSetUUID string      `json:"source_set_uuid"`
	ArtistID      int64       `json:"artist_id"`
	ArtistUUID    string      `json:"artist_uuid"`
	ShowUUID      string      `json:"show_uuid"`
	TrackPosition int64       `json:"track_position"`
	Duration      int64       `json:"duration"`
	Title         string      `json:"title"`
	Slug          string      `json:"slug"`
	Mp3URL        string      `json:"mp3_url"`
	Mp3Md5        string      `json:"mp3_md5"`
	FLACURL       interface{} `json:"flac_url"`
	FLACMd5       interface{} `json:"flac_md5"`
	UUID          string      `json:"uuid"`
	ID            int64       `json:"id"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
}

type Link struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

type ShowDetail struct {
	DisplayDate string   `json:"display_date"`
	Sources     []Source `json:"sources"`
}

type ArchiveMetadata struct {
	Metadata struct {
		Identifier string `json:"identifier"`
	} `json:"metadata"`
	Files []ArchiveFile `json:"files"`
}

type ArchiveFile struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Size   string `json:"size"`
	Title  string `json:"title"`
}

func main() {
	band := flag.String("band", "grateful-dead", "Band slug (e.g., grateful-dead)")
	year := flag.String("year", "", "Year to download (required)")
	outputDir := flag.String("output", "./downloads", "Output directory for downloads")
	format := flag.String("format", "mp3", "Preferred format: flac, mp3, or both")
	highestRated := flag.Bool("highest-rated", false, "Download only the highest rated source per show")
	flag.Parse()

	// Initialize logger with time-based log file
	var err error
	logger, err = NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("=== Dead-DL Started ===")
	logger.Info("Configuration: band=%s, year=%s, format=%s, output=%s, highest-rated=%v",
		*band, *year, *format, *outputDir, *highestRated)

	if *year == "" {
		logger.Fatal("Year is required. Use -year flag")
	}

	logger.Debug("Creating output directory: %s", *outputDir)
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logger.Fatal("Failed to create output directory %s: %v", *outputDir, err)
	}

	logger.Info("Fetching shows for %s in %s...", *band, *year)
	shows, err := fetchShows(*band, *year)
	if err != nil {
		logger.Fatal("Failed to fetch shows: %v", err)
	}

	logger.Info("Found %d shows for %s in %s", len(shows), *band, *year)
	logger.Println("") // Blank line for readability

	for i, show := range shows {
		logger.Printf("[%d/%d] Processing show: %s at %s, %s\n",
			i+1, len(shows), show.DisplayDate, show.Venue.Name, show.Venue.Location)

		// Fetch full show details which includes sources
		showDetail, err := fetchShowDetail(*band, show.DisplayDate)
		if err != nil {
			logger.Error("Failed to fetch show details for %s: %v", show.DisplayDate, err)
			continue
		}

		if len(showDetail.Sources) == 0 {
			logger.Printf("  No sources found for this show\n")
			continue
		}

		if len(showDetail.Sources) > 1 && *highestRated {
			// Select highest rated source
			bestSource := fetchHighestRatedSource(showDetail.Sources)
			if bestSource == nil {
				logger.Printf("  No valid sources found for this show\n")
				continue
			}
			showDetail.Sources = []Source{*bestSource}
			logger.Printf("  Selected highest rated source with avg rating %.2f\n", bestSource.AvgRating)
		}

		for j, source := range showDetail.Sources {
			logger.Printf("  Source [%d/%d]: ", j+1, len(showDetail.Sources))

			// Find archive.org link
			var archiveURL string
			for _, link := range source.Links {
				if strings.Contains(link.URL, "archive.org") {
					archiveURL = link.URL
					break
				}
			}

			if archiveURL == "" {
				logger.Println("No archive.org link found")
				continue
			}

			// Extract identifier from URL
			parts := strings.Split(archiveURL, "/")
			identifier := parts[len(parts)-1]
			logger.Printf("archive.org identifier: %s\n", identifier)

			// Create show directory
			showDir := filepath.Join(*outputDir, *band, *year, show.DisplayDate)
			if j > 0 {
				showDir = fmt.Sprintf("%s-source%d", showDir, j+1)
			}
			if err := os.MkdirAll(showDir, 0755); err != nil {
				logger.Error("Failed to create show directory: %v", err)
				continue
			}

			// Download files
			if err := downloadArchiveFiles(identifier, showDir, *format); err != nil {
				logger.Error("Failed to download files: %v", err)
				continue
			}

			logger.Printf("    ✓ Downloaded to %s\n", showDir)
		}
	}

	logger.Println("\nDownload complete!")
}

func fetchShows(band, year string) ([]Show, error) {
	url := fmt.Sprintf("%s/artists/%s/years/%s", RelistenAPIBase, band, year)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var showsResp ShowsResponse
	if err := json.NewDecoder(resp.Body).Decode(&showsResp); err != nil {
		return nil, err
	}

	return showsResp.Shows, nil
}

func fetchShowDetail(band, date string) (*ShowDetail, error) {
	url := fmt.Sprintf("%s/artists/%s/shows/%s", RelistenAPIBase, band, date)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var showDetail ShowDetail
	if err := json.NewDecoder(resp.Body).Decode(&showDetail); err != nil {
		return nil, err
	}

	return &showDetail, nil
}

func fetchHighestRatedSource(sources []Source) *Source {
	var bestSource *Source
	highestRating := 0.0
	for i, source := range sources {
		if source.AvgRating > highestRating {
			highestRating = source.AvgRating
			bestSource = &sources[i]
		}
	}
	return bestSource
}

func downloadArchiveFiles(identifier, outputDir, format string) error {
	// Fetch metadata
	url := fmt.Sprintf("%s/metadata/%s", ArchiveAPIBase, identifier)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("archive.org API returned status %d", resp.StatusCode)
	}

	var metadata ArchiveMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return err
	}

	// Filter files by format
	var filesToDownload []ArchiveFile
	wantFlac := format == "flac" || format == "both"
	wantMp3 := format == "mp3" || format == "both"

	for _, file := range metadata.Files {
		// Skip non-audio files
		if !isAudioFile(file.Name) {
			continue
		}

		fileNameLower := strings.ToLower(file.Name)
		fileFormat := strings.ToLower(file.Format)

		// Check if file matches requested format by extension (more reliable)
		isFlac := strings.HasSuffix(fileNameLower, ".flac")
		isMp3 := strings.HasSuffix(fileNameLower, ".mp3") ||
			strings.Contains(fileFormat, "mp3") ||
			fileFormat == "vbr mp3"

		// Also check format field for FLAC
		if !isFlac && strings.Contains(fileFormat, "flac") {
			isFlac = true
		}

		if (wantFlac && isFlac) || (wantMp3 && isMp3) {
			filesToDownload = append(filesToDownload, file)
		}
	}

	// If no files found and format was flac, try mp3 as fallback
	if len(filesToDownload) == 0 && format == "flac" {
		logger.Println("    - No FLAC files found, falling back to MP3...")
		wantMp3 = true
		wantFlac = false

		for _, file := range metadata.Files {
			if !isAudioFile(file.Name) {
				continue
			}

			fileNameLower := strings.ToLower(file.Name)
			fileFormat := strings.ToLower(file.Format)

			isMp3 := strings.HasSuffix(fileNameLower, ".mp3") ||
				strings.Contains(fileFormat, "mp3") ||
				fileFormat == "vbr mp3"

			if isMp3 {
				filesToDownload = append(filesToDownload, file)
			}
		}
	}

	if len(filesToDownload) == 0 {
		return fmt.Errorf("no audio files found in requested format")
	}

	// Download each file
	downloadErrors := []string{}
	successCount := 0

	for _, file := range filesToDownload {
		fileURL := fmt.Sprintf("%s/download/%s/%s", ArchiveAPIBase, identifier, file.Name)

		// Use title for filename if available, otherwise use original name
		fileName := file.Name
		if file.Title != "" {
			// Get extension from original filename
			ext := filepath.Ext(file.Name)
			// Sanitize title and use it as filename
			sanitizedTitle := sanitizeFilename(file.Title)
			fileName = sanitizedTitle + ext
		}

		filePath := filepath.Join(outputDir, fileName)

		// Check if file already exists and verify size
		if fileInfo, err := os.Stat(filePath); err == nil {
			// File exists, check if size matches
			localSize := fileInfo.Size()
			remoteSize, parseErr := parseFileSize(file.Size)

			if parseErr != nil {
				// Can't parse remote size, log warning and re-download
				logger.Printf("    - Re-downloading %s (unable to verify size: %v)\n", fileName, parseErr)
			} else if localSize == remoteSize {
				// Sizes match, skip download
				logger.Printf("    - Skipping %s (already exists, size: %d bytes)\n", fileName, localSize)
				successCount++
				continue
			} else {
				// Sizes don't match, re-download
				logger.Printf("    - Re-downloading %s (size mismatch: local=%d, remote=%d)\n", fileName, localSize, remoteSize)
			}
		}

		logger.Printf("    - Downloading %s...\n", fileName)
		if err := downloadFile(fileURL, filePath, fileName); err != nil {
			// Handle specific HTTP error codes
			if strings.Contains(err.Error(), "status 401") {
				logger.Printf("    - ⚠ Skipping %s (restricted/requires authentication)\n", fileName)
				downloadErrors = append(downloadErrors, fmt.Sprintf("%s: restricted", fileName))
			} else if strings.Contains(err.Error(), "status 403") {
				logger.Printf("    - ⚠ Skipping %s (forbidden/restricted)\n", fileName)
				downloadErrors = append(downloadErrors, fmt.Sprintf("%s: forbidden", fileName))
			} else {
				logger.Printf("    - ✗ Failed to download %s: %v\n", fileName, err)
				downloadErrors = append(downloadErrors, fmt.Sprintf("%s: %v", fileName, err))
			}
			continue
		}

		logger.Printf("    - ✓ Downloaded %s\n", fileName)
		successCount++
		time.Sleep(100 * time.Millisecond) // Be nice to the server
	}

	// Return error only if all downloads failed
	if successCount == 0 && len(downloadErrors) > 0 {
		return fmt.Errorf("all downloads failed: %s", strings.Join(downloadErrors, "; "))
	}

	// Log warnings if some downloads failed
	if len(downloadErrors) > 0 && successCount > 0 {
		logger.Printf("    - ⚠ %d file(s) failed to download (see above)\n", len(downloadErrors))
	}

	return nil
}

func isAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	audioExts := []string{".flac", ".mp3", ".ogg", ".shn", ".wav", ".m4a"}
	for _, ext2 := range audioExts {
		if ext == ext2 {
			return true
		}
	}
	return false
}

// parseFileSize converts the archive.org size string to int64
// The size field is typically a string representation of bytes
func parseFileSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Try to parse as int64 directly (bytes)
	var size int64
	_, err := fmt.Sscanf(sizeStr, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size: %w", err)
	}

	return size, nil
}

func sanitizeFilename(name string) string {
	// Replace invalid filename characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Remove leading/trailing spaces and dots
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")
	// Replace multiple spaces/underscores with single underscore
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	// Limit length to avoid filesystem issues
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}

func downloadFile(url, filepath, displayName string) error {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Create output file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get content length for progress bar
	contentLength := resp.ContentLength

	// Create progress bar
	var bar *progressbar.ProgressBar
	if contentLength > 0 {
		// Content length is known, show byte progress
		bar = progressbar.DefaultBytes(
			contentLength,
			fmt.Sprintf("      %s", displayName),
		)
	} else {
		// Content length unknown, show indeterminate progress
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(fmt.Sprintf("      %s", displayName)),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(50),
		)
	}

	// Create multi-writer to write to both file and progress bar
	writer := io.MultiWriter(out, bar)

	// Copy data to file and update progress bar
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	// Finish the progress bar
	bar.Finish()

	return nil
}
