package scramble

import (
	"context"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/streaming-service/models"
	"gofr.dev/pkg/gofr/container"
	"os"
	"strings"
	"sync"
	"time"
)

type Scrambler struct {
	processor Processor
	sql       container.DB
	opts      *ScramblerOptions
}

type ScramblerOptions struct {
	WorkerPoolSize int
	MaxBatchSize   int
	HLSFolder      string
	Encryption     Encryption
}

type videoSessionResult struct {
	index        int
	videoSession *models.VideoSession
	newLine      string
	err          error
}

func NewScrambler(processor Processor, sql container.DB, options *ScramblerOptions) *Scrambler {
	return &Scrambler{
		processor: processor,
		sql:       sql,
		opts:      options,
	}
}

func (s *Scrambler) Scramble(m3u8File string, videoId string, userId string) ([]byte, error) {
	start := time.Now()

	fileContents, err := os.ReadFile(m3u8File)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(fileContents), "\n")

	workerChan := make(chan struct{}, s.opts.WorkerPoolSize)
	resultChan := make(chan videoSessionResult, len(lines))
	batchInsertChan := make(chan []*models.VideoSession)
	var wg sync.WaitGroup
	var lineMu sync.Mutex

	for i, line := range lines {
		if !strings.HasPrefix(line, "playlist") {
			continue
		}

		wg.Add(1)
		workerChan <- struct{}{}

		go func(index int, fragmentLine string) {
			defer wg.Done()
			defer func() { <-workerChan }()

			session, newLine, err := s.processor.ProcessSegment(videoId, userId, fragmentLine, s.opts.HLSFolder)
			resultChan <- videoSessionResult{index, session, newLine, err}
		}(i, line)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	go s.handleResults(resultChan, lines, batchInsertChan, &lineMu)
	go s.batchInsert(batchInsertChan)

	wg.Wait()

	finalContent := strings.Join(lines, "\n")
	finalContent, err = s.opts.Encryption.Encrypt(finalContent)
	if err != nil {
		fmt.Println("err on encoding", err.Error())
		return nil, err
	}

	fmt.Println("Scramble took:", time.Since(start))
	return []byte(finalContent), nil
}

func (s *Scrambler) handleResults(resultChan <-chan videoSessionResult, lines []string, batchInsertChan chan<- []*models.VideoSession, lineMu *sync.Mutex) {
	var batch []*models.VideoSession

	for result := range resultChan {
		if result.err != nil {
			fmt.Println("result err at index", result.index, result.err.Error())
			continue
		}

		lineMu.Lock()
		lines[result.index] = result.newLine
		lineMu.Unlock()

		batch = append(batch, result.videoSession)

		if len(batch) >= s.opts.MaxBatchSize {
			batchInsertChan <- batch
			batch = nil
		}
	}

	if len(batch) > 0 {
		batchInsertChan <- batch
	}
	close(batchInsertChan)
}

func (s *Scrambler) batchInsert(batchInsertChan <-chan []*models.VideoSession) {
	for batch := range batchInsertChan {
		if err := s.insertBatch(batch); err != nil {
			fmt.Println("Batch Insert Error: ", err)
		}
	}
}

func (s *Scrambler) insertBatch(videoSessions []*models.VideoSession) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var valueStrings []string
	var valueArgs []interface{}

	for _, session := range videoSessions {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, session.Id, session.UserId, session.VideoId, session.FragmentHash, session.FragmentPath, session.Token)
	}

	query := fmt.Sprintf("INSERT INTO video_sessions (id, user_id, video_id, fragment_hash, fragment_path, token) VALUES %s", strings.Join(valueStrings, ","))
	_, err := s.sql.ExecContext(ctx, query, valueArgs...)
	return err
}
