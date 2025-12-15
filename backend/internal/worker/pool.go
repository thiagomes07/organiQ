// internal/worker/pool.go
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"organiq/internal/domain/repository"
	"organiq/internal/infra/ai"
	"organiq/internal/infra/queue"
	"organiq/internal/util"
)

// WorkerPool gerencia múltiplos workers consumindo das filas
type WorkerPool struct {
	generatorWorkers []*ArticleGeneratorWorker
	publisherWorkers []*ArticlePublisherWorker
	workerCount      int
	mu               sync.RWMutex
	wg               sync.WaitGroup
	shutdownChan     chan struct{}
}

// NewWorkerPool cria nova instância do pool de workers
func NewWorkerPool(
	workerCount int,
	queueService queue.QueueService,
	articleJobRepo repository.ArticleJobRepository,
	articleIdeaRepo repository.ArticleIdeaRepository,
	articleRepo repository.ArticleRepository,
	businessRepo repository.BusinessRepository,
	integrationRepo repository.IntegrationRepository,
	agentClient *ai.AgentClient,
	cryptoService *util.CryptoService,
	pollInterval time.Duration,
	maxRetries int,
) *WorkerPool {
	generatorWorkers := make([]*ArticleGeneratorWorker, workerCount)
	publisherWorkers := make([]*ArticlePublisherWorker, workerCount)

	for i := 0; i < workerCount; i++ {
		generatorWorkers[i] = NewArticleGeneratorWorker(
			queueService,
			articleJobRepo,
			articleIdeaRepo,
			businessRepo,
			agentClient,
			pollInterval,
			maxRetries,
		)

		publisherWorkers[i] = NewArticlePublisherWorker(
			queueService,
			articleRepo,
			businessRepo,
			integrationRepo,
			agentClient,
			cryptoService,
			pollInterval,
			maxRetries,
		)
	}

	return &WorkerPool{
		generatorWorkers: generatorWorkers,
		publisherWorkers: publisherWorkers,
		workerCount:      workerCount,
		shutdownChan:     make(chan struct{}),
	}
}

// Start inicia todos os workers em goroutines separadas
// Retorna um canal que pode ser usado para aguardar shutdown completo
func (p *WorkerPool) Start(ctx context.Context) <-chan struct{} {
	log.Info().
		Int("worker_count", p.workerCount).
		Msg("WorkerPool iniciando")

	p.mu.Lock()
	defer p.mu.Unlock()

	shutdownDone := make(chan struct{})

	// Iniciar generator workers em goroutines separadas
	for i, worker := range p.generatorWorkers {
		p.wg.Add(1)

		go func(index int, w *ArticleGeneratorWorker) {
			defer p.wg.Done()

			log.Info().
				Int("worker_index", index).
				Str("worker_id", w.workerID).
				Str("type", "generator").
				Msg("Generator worker iniciado")

			if err := w.Start(ctx); err != nil && err != context.Canceled {
				log.Error().
					Err(err).
					Int("worker_index", index).
					Str("worker_id", w.workerID).
					Msg("Generator worker parou com erro")
			} else {
				log.Info().
					Int("worker_index", index).
					Str("worker_id", w.workerID).
					Msg("Generator worker parou gracefully")
			}
		}(i, worker)
	}

	// Iniciar publisher workers em goroutines separadas
	for i, worker := range p.publisherWorkers {
		p.wg.Add(1)

		go func(index int, w *ArticlePublisherWorker) {
			defer p.wg.Done()

			log.Info().
				Int("worker_index", index).
				Str("worker_id", w.workerID).
				Str("type", "publisher").
				Msg("Publisher worker iniciado")

			if err := w.Start(ctx); err != nil && err != context.Canceled {
				log.Error().
					Err(err).
					Int("worker_index", index).
					Str("worker_id", w.workerID).
					Msg("Publisher worker parou com erro")
			} else {
				log.Info().
					Int("worker_index", index).
					Str("worker_id", w.workerID).
					Msg("Publisher worker parou gracefully")
			}
		}(i, worker)
	}

	// Goroutine que aguarda todos os workers terminarem
	go func() {
		p.wg.Wait()
		log.Info().Msg("WorkerPool shutdown completo")
		close(shutdownDone)
	}()

	return shutdownDone
}

// Shutdown aguarda graceful shutdown de todos os workers
// Cancela o contexto e aguarda que todos terminem
func (p *WorkerPool) Shutdown(ctx context.Context) error {
	log.Info().
		Int("worker_count", p.workerCount).
		Msg("WorkerPool iniciando shutdown")

	// Aguardar que todos os workers completem
	// Se demorar muito, retornar timeout
	done := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("WorkerPool shutdown bem-sucedido")
		return nil

	case <-ctx.Done():
		log.Warn().Msg("WorkerPool shutdown timeout")
		return ctx.Err()
	}
}

// GetWorkerCount retorna número de workers no pool
func (p *WorkerPool) GetWorkerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.workerCount
}
