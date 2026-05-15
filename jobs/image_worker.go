package jobs

import (
	"api/logger"
	"api/metrics"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"sync"
	"time"

	_ "image/gif"
)

// ImageJob representa uma tarefa de processamento de imagem
type ImageJob struct {
	ProductID  int
	FileName   string
	FileData   []byte
	TargetSize int // tamanho máximo em bytes, 0 = sem limite
	OnComplete func(processedData []byte, err error)
}

// ImageWorkerPool gerencia um pool de workers para processar imagens
type ImageWorkerPool struct {
	jobs       chan ImageJob
	wg         sync.WaitGroup
	numWorkers int
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewImageWorkerPool cria um novo worker pool com N workers
func NewImageWorkerPool(numWorkers int) *ImageWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ImageWorkerPool{
		jobs:       make(chan ImageJob, 100), // buffer de 100 jobs
		numWorkers: numWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Iniciar workers
	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	logger.Log.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"Image worker pool started",
		slog.Int("workers", numWorkers),
	)

	return pool
}

// Submit enfileira um job de processamento de imagem
func (p *ImageWorkerPool) Submit(job ImageJob) error {
	select {
	case p.jobs <- job:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool está fechado")
	default:
		// Fila cheia, retornar erro em vez de bloquear
		return fmt.Errorf("fila de processamento cheia, tente novamente")
	}
}

// worker processa jobs da fila
func (p *ImageWorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			logger.Log.LogAttrs(
				p.ctx,
				slog.LevelInfo,
				"Image worker stopped",
				slog.Int("worker_id", id),
			)
			return

		case job, ok := <-p.jobs:
			if !ok {
				// Canal fechado
				return
			}

			start := time.Now()
			processedData, err := processImageJob(job)
			duration := time.Since(start)

			if err != nil {
				logger.Log.LogAttrs(
					p.ctx,
					slog.LevelWarn,
					"Image processing failed",
					slog.Int("worker_id", id),
					slog.Int("product_id", job.ProductID),
					slog.String("file_name", job.FileName),
					slog.String("error", err.Error()),
					slog.Int64("duration_ms", duration.Milliseconds()),
				)
			} else {
				logger.Log.LogAttrs(
					p.ctx,
					slog.LevelInfo,
					"Image processed successfully",
					slog.Int("worker_id", id),
					slog.Int("product_id", job.ProductID),
					slog.String("file_name", job.FileName),
					slog.Int("original_size", len(job.FileData)),
					slog.Int("processed_size", len(processedData)),
					slog.Int64("duration_ms", duration.Milliseconds()),
				)
				metrics.ImageUploadsTotal.Inc()
			}

			// Chamar callback
			if job.OnComplete != nil {
				job.OnComplete(processedData, err)
			}
		}
	}
}

// Close encerra o worker pool gracefully
func (p *ImageWorkerPool) Close() error {
	p.cancel()
	close(p.jobs)
	p.wg.Wait()

	logger.Log.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"Image worker pool closed",
		slog.Int("workers", p.numWorkers),
	)

	return nil
}

// processImageJob processa uma imagem individual
// - Valida formato
// - Comprime se necessário
// - Retorna dados processados
func processImageJob(job ImageJob) ([]byte, error) {
	if len(job.FileData) == 0 {
		return nil, fmt.Errorf("arquivo vazio")
	}

	// 1. Decodificar imagem para validar e detectar tipo real
	img, format, err := image.DecodeConfig(bytes.NewReader(job.FileData))
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar imagem: %w", err)
	}

	logger.Log.LogAttrs(
		context.Background(),
		slog.LevelDebug,
		"Image decoded",
		slog.String("format", format),
		slog.Int("width", img.Width),
		slog.Int("height", img.Height),
	)

	// 2. Validar dimensões
	if img.Width > 8000 || img.Height > 8000 {
		return nil, fmt.Errorf("imagem muito grande: %dx%d (máximo 4000x4000)", img.Width, img.Height)
	}

	if img.Width < 100 || img.Height < 100 {
		return nil, fmt.Errorf("imagem muito pequena: %dx%d (mínimo 100x100)", img.Width, img.Height)
	}

	// 3. Decodificar imagem completa
	fullImg, _, err := image.Decode(bytes.NewReader(job.FileData))
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar imagem completa: %w", err)
	}

	// 4. Comprimir para JPEG
	processedBuffer := &bytes.Buffer{}
	err = jpeg.Encode(processedBuffer, fullImg, &jpeg.Options{
		Quality: 85, // Balance entre qualidade e tamanho
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao comprimir imagem: %w", err)
	}

	processedData := processedBuffer.Bytes()

	// 5. Validar tamanho máximo
	if job.TargetSize > 0 && len(processedData) > job.TargetSize {
		return nil, fmt.Errorf(
			"imagem comprimida ainda acima do limite: %d bytes (máximo %d)",
			len(processedData),
			job.TargetSize,
		)
	}

	return processedData, nil
}

// ImageProcessor é um helper que enfileira jobs
type ImageProcessor struct {
	pool *ImageWorkerPool
}

// NewImageProcessor cria um novo processador
func NewImageProcessor(numWorkers int) *ImageProcessor {
	return &ImageProcessor{
		pool: NewImageWorkerPool(numWorkers),
	}
}

// ProcessAsync enfileira uma imagem para processamento assíncrono
func (ip *ImageProcessor) ProcessAsync(job ImageJob) error {
	return ip.pool.Submit(job)
}

// ProcessSync processa uma imagem sincronamente (bloqueia)
func ProcessImageSync(fileData []byte) ([]byte, error) {
	job := ImageJob{
		FileData:   fileData,
		TargetSize: 5 * 1024 * 1024, // 5MB max
	}
	return processImageJob(job)
}

// Shutdown encerra o processador
func (ip *ImageProcessor) Shutdown() error {
	return ip.pool.Close()
}
