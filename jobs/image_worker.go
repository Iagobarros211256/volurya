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
	_ "image/png"
)

// ImageJob representa uma tarefa de processamento de imagem
type ImageJob struct {
	ProductID          int
	FileName           string
	FileData           []byte
	TargetSize         int // tamanho máximo em bytes, 0 = sem limite
	ResizeWidth        int // largura alvo para redimensionamento, 0 = não redimensionar
	ResizeHeight       int // altura alvo para redimensionamento, 0 = não redimensionar
	CompressionQuality int // 1-100, qualidade JPEG (padrão 85)
	OnComplete         func(processedData []byte, err error)
	StorageCallback    func([]byte) (string, error) // callback para upload em R2
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

// processImageJob processa uma imagem seguindo o pipeline:
// 1. Valida formato e dimensões
// 2. Redimensiona se necessário
// 3. Comprime para JPEG
// 4. Chama callback de armazenamento (R2)
func processImageJob(job ImageJob) ([]byte, error) {
	ctx := context.Background()

	// ===== STAGE 1: Validate =====
	stageStart := time.Now()
	if len(job.FileData) == 0 {
		return nil, fmt.Errorf("arquivo vazio")
	}

	// Decodificar config para validar
	imgConfig, format, err := image.DecodeConfig(bytes.NewReader(job.FileData))
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar imagem: %w", err)
	}

	logger.Log.LogAttrs(
		ctx,
		slog.LevelDebug,
		"Pipeline stage: VALIDATE",
		slog.String("format", format),
		slog.Int("width", imgConfig.Width),
		slog.Int("height", imgConfig.Height),
		slog.Int64("duration_ms", time.Since(stageStart).Milliseconds()),
	)

	// Validar dimensões
	if imgConfig.Width > 8000 || imgConfig.Height > 8000 {
		return nil, fmt.Errorf("imagem muito grande: %dx%d (máximo 8000x8000)", imgConfig.Width, imgConfig.Height)
	}

	if imgConfig.Width < 100 || imgConfig.Height < 100 {
		return nil, fmt.Errorf("imagem muito pequena: %dx%d (mínimo 100x100)", imgConfig.Width, imgConfig.Height)
	}

	// Decodificar imagem completa
	fullImg, _, err := image.Decode(bytes.NewReader(job.FileData))
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar imagem completa: %w", err)
	}

	// ===== STAGE 2: Resize (se necessário) =====
	var processedImg image.Image
	if job.ResizeWidth > 0 && job.ResizeHeight > 0 {
		stageStart = time.Now()
		processedImg = resizeImageAspectRatio(fullImg, job.ResizeWidth, job.ResizeHeight)
		logger.Log.LogAttrs(
			ctx,
			slog.LevelDebug,
			"Pipeline stage: RESIZE",
			slog.Int("target_width", job.ResizeWidth),
			slog.Int("target_height", job.ResizeHeight),
			slog.Int("actual_width", processedImg.Bounds().Dx()),
			slog.Int("actual_height", processedImg.Bounds().Dy()),
			slog.Int64("duration_ms", time.Since(stageStart).Milliseconds()),
		)
	} else {
		processedImg = fullImg
	}

	// ===== STAGE 3: Compress =====
	stageStart = time.Now()
	quality := job.CompressionQuality
	if quality == 0 {
		quality = 85 // padrão
	}
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	processedBuffer := &bytes.Buffer{}
	err = jpeg.Encode(processedBuffer, processedImg, &jpeg.Options{
		Quality: quality,
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao comprimir imagem: %w", err)
	}

	processedData := processedBuffer.Bytes()
	logger.Log.LogAttrs(
		ctx,
		slog.LevelDebug,
		"Pipeline stage: COMPRESS",
		slog.Int("original_size", len(job.FileData)),
		slog.Int("compressed_size", len(processedData)),
		slog.Int("compression_quality", quality),
		slog.Int64("duration_ms", time.Since(stageStart).Milliseconds()),
	)

	// Validar tamanho máximo
	if job.TargetSize > 0 && len(processedData) > job.TargetSize {
		return nil, fmt.Errorf(
			"imagem comprimida ainda acima do limite: %d bytes (máximo %d)",
			len(processedData),
			job.TargetSize,
		)
	}

	// ===== STAGE 4: Storage Callback (R2 Upload) =====
	if job.StorageCallback != nil {
		stageStart = time.Now()
		_, err := job.StorageCallback(processedData)
		if err != nil {
			return nil, fmt.Errorf("falha ao fazer upload para storage: %w", err)
		}
		logger.Log.LogAttrs(
			ctx,
			slog.LevelDebug,
			"Pipeline stage: STORAGE",
			slog.Int("size", len(processedData)),
			slog.Int64("duration_ms", time.Since(stageStart).Milliseconds()),
		)
	}

	return processedData, nil
}

// resizeImageAspectRatio redimensiona a imagem mantendo a proporção de aspecto
// Usa interpolação linear (bilinear) para melhor qualidade
func resizeImageAspectRatio(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Calcular scaling factor mantendo aspecto
	scale := 1.0
	if srcWidth > maxWidth {
		scale = float64(maxWidth) / float64(srcWidth)
	}
	if srcHeight > maxHeight {
		newScale := float64(maxHeight) / float64(srcHeight)
		if newScale < scale {
			scale = newScale
		}
	}

	// Se não precisa redimensionar, retorna original
	if scale >= 1.0 {
		return src
	}

	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// Criar imagem redimensionada com interpolação nearest-neighbor
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simples nearest-neighbor scaling
	xRatio := float64(srcWidth-1) / float64(newWidth-1)
	yRatio := float64(srcHeight-1) / float64(newHeight-1)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			px := int(float64(x) * xRatio)
			py := int(float64(y) * yRatio)
			dst.Set(x, y, src.At(bounds.Min.X+px, bounds.Min.Y+py))
		}
	}

	return dst
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
// Parâmetros padrão: redimensiona para 1200x1200 max, qualidade 85
func ProcessImageSync(fileData []byte) ([]byte, error) {
	job := ImageJob{
		FileData:           fileData,
		TargetSize:         5 * 1024 * 1024, // 5MB max
		ResizeWidth:        1200,
		ResizeHeight:       1200,
		CompressionQuality: 85,
	}
	return processImageJob(job)
}

// Shutdown encerra o processador
func (ip *ImageProcessor) Shutdown() error {
	return ip.pool.Close()
}

/*


🔴 Race condition no Close()
gofunc (p *ImageWorkerPool) Close() error {
    p.cancel()      // cancela context
    close(p.jobs)   // fecha canal
    p.wg.Wait()
p.cancel() e close(p.jobs) disparados juntos causam race condition — um worker pode estar no select e receber tanto ctx.Done() quanto um job do canal fechado simultaneamente. O correto é cancelar o context e deixar os workers drenarem naturalmente, sem fechar o canal:
gofunc (p *ImageWorkerPool) Close() error {
    p.cancel()
    p.wg.Wait()  // workers param quando ctx.Done() é recebido
    return nil
}

🔴 close(p.jobs) com jobs pendentes na fila
Se houver 50 jobs na fila e Close() for chamado, os jobs pendentes são descartados silenciosamente. Workers que ainda estão processando vão tentar ler do canal fechado. Decida explicitamente: drena a fila ou descarta:
go// Opção 1: drena fila antes de fechar
// Opção 2: documenta que jobs pendentes são descartados

🔴 resizeImageAspectRatio usa nearest-neighbor mas comenta bilinear
go// Usa interpolação linear (bilinear) para melhor qualidade
func resizeImageAspectRatio(...) {
    // Simples nearest-neighbor scaling  ← contradiz o comentário
Nearest-neighbor produz imagens pixeladas em downscaling. O comentário promete bilinear mas a implementação é nearest-neighbor. Para imagens de produto isso é qualidade visivelmente inferior. Use golang.org/x/image/draw:
goimport "golang.org/x/image/draw"

dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

🔴 Divisão por zero em resizeImageAspectRatio
goxRatio := float64(srcWidth-1) / float64(newWidth-1)
yRatio := float64(srcHeight-1) / float64(newHeight-1)
Se newWidth == 1 ou newHeight == 1, divisão por zero resulta em +Inf. Adicione guard:
goif newWidth <= 1 || newHeight <= 1 {
    return src
}

🟡 StorageCallback no job cria acoplamento duplo
O product_controller.go já tem um OnComplete callback que faz o upload para R2. StorageCallback é um segundo mecanismo para a mesma coisa — dois callbacks para storage é confuso. Escolha um padrão.

🟡 Buffer de 100 jobs hardcoded
gojobs: make(chan ImageJob, 100)
Para uma loja pequena é suficiente, mas deveria ser configurável:
gofunc NewImageWorkerPool(numWorkers, bufferSize int) *ImageWorkerPool {

🟡 processImageJob decodifica a imagem duas vezes
goimgConfig, format, err := image.DecodeConfig(bytes.NewReader(job.FileData))
// ...
fullImg, _, err := image.Decode(bytes.NewReader(job.FileData))
DecodeConfig lê apenas o header. Decode lê o arquivo completo. Duas leituras do mesmo buffer. Dá pra validar dimensões depois de Decode:
gofullImg, format, err := image.Decode(bytes.NewReader(job.FileData))
bounds := fullImg.Bounds()
if bounds.Dx() > 8000 || bounds.Dy() > 8000 { ... }

🟡 TargetSize validado após compressão sem retry
goif job.TargetSize > 0 && len(processedData) > job.TargetSize {
    return nil, fmt.Errorf("imagem comprimida ainda acima do limite...")
}
Se a imagem comprimida com qualidade 85 ainda excede o limite, retorna erro. Poderia tentar qualidade menor automaticamente:
gofor quality > 10 && len(processedData) > job.TargetSize {
    quality -= 10
    // recomprimir
}

🟢 ProcessImageSync exportada mas processImageJob não
gofunc ProcessImageSync(fileData []byte) ([]byte, error) { ... }  // exportada
func processImageJob(job ImageJob) ([]byte, error) { ... }       // não exportada
ProcessImageSync usa valores hardcoded (1200x1200, 85%) sem possibilidade de customização. Se for uma API pública do pacote, deveria aceitar parâmetros ou um ImageJob.

🟢 ctx local criado mas nunca usado para cancelamento
goctx := context.Background()
// passado apenas para logs, nunca para operações canceláveis
O context nos logs é correto, mas nenhuma operação dentro de processImageJob respeita cancelamento. Se o worker for encerrado durante o processamento, a função continua até terminar.

*/
