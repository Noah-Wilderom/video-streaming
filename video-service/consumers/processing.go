package consumers

import (
	"context"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/video-service/models"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gofr.dev/pkg/gofr"
	"os"
	"path/filepath"
	"time"
)

func ProcessVideo(c *gofr.Context) error {
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

	//probeCtx, probeCancel := context.WithTimeout(context.Background(), 15*time.Second)
	//defer probeCancel()
	//
	//metadata, err := ffprobe.ProbeURL(probeCtx, inputPath)
	//if err != nil {
	//	return err
	//}

	if err := os.Mkdir(fmt.Sprintf("%s/hls", video.Path), 0x777); err != nil {
		c.Logger.Error("creating a hls directory failed", err)
		return err
	}

	err = ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{"profile:v": "high10", "c:a": "aac", "c:v": "libx264", "level": "3.0", "start_number": "0", "hls_time": "10", "hls_list_size": "0", "f": "hls"}).
		ErrorToStdOut().
		Run()

	if err != nil {
		err = fmt.Errorf("error transcoding video: %v", err)
		c.Logger.Error(err)
		return err
	}

	//bitrate, err := strconv.Atoi(metadata.Format.BitRate)
	//if err != nil {
	//	return err
	//}

	//videoMetadata := &models.Metadata{
	//	Resolution: fmt.Sprintf("%dx%d", metadata.Streams[0].Width, metadata.Streams[0].Height),
	//	Duration:   int(metadata.Format.Duration().Seconds()),
	//	Format:     metadata.Format.FormatName,
	//	Codec:      metadata.Streams[0].CodecName,
	//	Bitrate:    bitrate,
	//}
	//
	//videoMetadataJson, err := json.Marshal(videoMetadata)
	//if err != nil {
	//	return err
	//}

	//video.Metadata = string(videoMetadataJson)

	updateCtx, updateCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer updateCancel()

	_, err = c.SQL.ExecContext(updateCtx, "UPDATE videos SET metadata = ?, status = ?, processed_at = ? WHERE id = ?", video.Metadata, "processed", time.Now(), video.Id)
	if err != nil {
		return err
	}

	c.Logger.Info("Processed video", video)

	return nil
}
