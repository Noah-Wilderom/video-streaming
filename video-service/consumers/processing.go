package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/video-service/models"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gofr.dev/pkg/gofr"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ProcessVideo(c *gofr.Context) error {
	const (
		previewInSeconds = 20
	)

	c.Logger.Info("ProcessVideo function called")
	var video *models.Video

	err := c.Bind(&video)
	if err != nil {
		c.Logger.Error(err)

		// returning nil here as we would like to ignore the
		// incompatible message and continue reading forward
		return nil
	}

	c.Logger.Info("Received video for processing ", video)
	inputPath := fmt.Sprintf("%s/original.mp4", video.Path)
	outputPath := filepath.Join(video.Path, "hls", "playlist.m3u8")
	outputPathPreview := filepath.Join(video.Path, "hls_preview", "playlist.m3u8")

	if err := os.Mkdir(fmt.Sprintf("%s/hls", video.Path), 0x777); err != nil {
		c.Logger.Error("creating a hls directory failed", err)
		return err
	}

	if err := os.Mkdir(fmt.Sprintf("%s/hls_preview", video.Path), 0x777); err != nil {
		c.Logger.Error("creating a hls_preview directory failed", err)
		return err
	}

	err = ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"profile:v":     "high10",
			"c:a":           "copy", // No re-encoding
			"c:v":           "libx264",
			"start_number":  "0",
			"hls_time":      "10",
			"hls_list_size": "0",
			"f":             "hls",
			"threads":       "auto",
		}).
		ErrorToStdOut().
		Run()
	if err != nil {
		err = fmt.Errorf("error transcoding video: %v", err)
		c.Logger.Error(err)
		return err
	}

	c.Logger.Info("extracting first frame as image")
	err = ffmpeg.Input(inputPath).
		Output(filepath.Join(video.Path, "first_frame.jpg"), ffmpeg.KwArgs{"frames:v": 1, "q:v": 2}).
		ErrorToStdOut().
		Run()
	if err != nil {
		err = fmt.Errorf("error extracting first frame as image: %v", err)
		c.Logger.Error(err)
		return err
	}

	err = ffmpeg.Input(inputPath).
		Output(outputPathPreview, ffmpeg.KwArgs{
			"t":             previewInSeconds, // Limit duration
			"c:a":           "aac",            // Re-encode audio to AAC
			"c:v":           "libx264",        // Re-encode video using libx264
			"profile:v":     "high",           // Use high profile for 8-bit video
			"pix_fmt":       "yuv420p",        // Convert to 8-bit format
			"crf":           "23",             // Set quality level
			"maxrate":       "5M",             // Maximum bitrate
			"bufsize":       "10M",            // Buffer size
			"preset":        "medium",         // Control speed vs compression ratio
			"start_number":  "0",
			"hls_time":      "10",
			"hls_list_size": "0",
			"f":             "hls",
			"threads":       "auto",
		}).
		ErrorToStdOut().
		Run()
	if err != nil {
		err = fmt.Errorf("error transcoding preview video: %v", err)
		c.Logger.Error(err)
		return err
	}

	videoMetadata, err := getFFprobeMetadata(inputPath)
	if err != nil {
		return err
	}

	videoMetadataJson, err := json.Marshal(videoMetadata)
	if err != nil {
		return err
	}

	video.Metadata = string(videoMetadataJson)

	updateCtx, updateCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer updateCancel()

	_, err = c.SQL.ExecContext(updateCtx, "UPDATE videos SET metadata = ?, status = ?, processed_at = ? WHERE id = ?", video.Metadata, "processed", time.Now(), video.Id)
	if err != nil {
		return err
	}

	c.Logger.Info("Processed video", video)

	return nil
}

func getFFprobeMetadata(inputPath string) (*models.Metadata, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_format", "-show_streams", "-of", "json", inputPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, err
	}

	streams, ok := metadata["streams"].([]interface{})
	if !ok || len(streams) == 0 {
		return nil, fmt.Errorf("no streams found in metadata")
	}

	videoStream := streams[0].(map[string]interface{})

	// Helper function to get float64 value with nil check
	getFloat64 := func(key string, m map[string]interface{}) (float64, bool) {
		if val, exists := m[key]; exists {
			if floatVal, ok := val.(float64); ok {
				return floatVal, true
			}
		}
		return 0, false
	}

	// Extract resolution, duration, format, codec, and bitrate
	width, _ := getFloat64("width", videoStream)
	height, _ := getFloat64("height", videoStream)
	duration, _ := getFloat64("duration", metadata["format"].(map[string]interface{}))
	format := metadata["format"].(map[string]interface{})["format_name"].(string)
	codec := videoStream["codec_name"].(string)
	bitrate, _ := getFloat64("bit_rate", metadata["format"].(map[string]interface{}))

	videoMetadata := &models.Metadata{
		Resolution: fmt.Sprintf("%dx%d", int(width), int(height)),
		Duration:   int(duration),
		Format:     format,
		Codec:      codec,
		Bitrate:    int(bitrate),
	}

	return videoMetadata, nil
}
