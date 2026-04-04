package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var IntentExecutorService = &intentExecutorService{}

type intentExecutorService struct{}

func (s *intentExecutorService) ProcessPending(ctx context.Context, workerID string, limit int, leaseDuration time.Duration) (int, error) {
	ctx, span := StartServiceSpan(ctx, "IntentExecutor.ProcessPending",
		trace.WithAttributes(
			attribute.String("worker.id", workerID),
			attribute.Int("worker.batch_size", limit),
			attribute.Int64("worker.lease_seconds", int64(leaseDuration/time.Second)),
		),
	)
	defer span.End()

	processed := 0
	for range limitOrOne(limit) {
		intent, err := IntentService.ClaimNextPending(ctx, workerID, leaseDuration)
		if errors.Is(err, ErrIntentNotFound) {
			break
		}
		if err != nil {
			return processed, err
		}

		intentCtx, intentSpan := StartServiceSpan(ctx, "IntentExecutor.Execute",
			trace.WithAttributes(
				attribute.String("intent.id", intent.ID.Hex()),
				attribute.String("intent.kind", string(intent.Kind)),
				attribute.String("resource.id", intent.ResourceID.Hex()),
			),
		)

		externalRef, err := s.executeIntent(intentCtx, intent)
		if err != nil {
			logging.LoggerWithContext(intentCtx).Error("intent execution failed",
				zap.String("intent_id", intent.ID.Hex()),
				zap.String("kind", string(intent.Kind)),
				zap.Error(err),
			)
			intentSpan.End()
			_ = IntentService.MarkFailed(intentCtx, intent.ID, err.Error())
			continue
		}

		_ = IntentService.MarkSubmitted(intentCtx, intent.ID, externalRef, "submitted to executor")
		intentSpan.End()
		processed++
	}

	return processed, nil
}

func (s *intentExecutorService) executeIntent(ctx context.Context, intent *model.Intent) (string, error) {
	switch intent.Kind {
	case model.IntentKindBuild:
		if err := ManifestService.DispatchBuild(ctx, intent.ResourceID); err != nil {
			return intent.ExternalRef, err
		}
		manifest, err := ManifestService.Get(ctx, intent.ResourceID)
		if err != nil {
			return intent.ExternalRef, err
		}
		return manifest.PipelineID, nil
	case model.IntentKindRelease:
		if err := JobService.DispatchRelease(ctx, intent.ResourceID); err != nil {
			return intent.ExternalRef, err
		}
		job, err := JobService.Get(ctx, intent.ResourceID)
		if err != nil {
			return intent.ExternalRef, err
		}
		return fmt.Sprintf("argocd/%s", job.ApplicationName), nil
	default:
		return intent.ExternalRef, fmt.Errorf("unsupported intent kind %q", intent.Kind)
	}
}

func limitOrOne(limit int) []struct{} {
	if limit <= 0 {
		limit = 1
	}
	return make([]struct{}, limit)
}
