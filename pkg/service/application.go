package service

import (
	"context"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var ApplicationService = NewApplicationService()

type applicationService struct{}

func NewApplicationService() *applicationService {
	return &applicationService{}
}

func (s *applicationService) Get(ctx context.Context, id primitive.ObjectID) (*model.Application, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "get_application"),
		zap.String("application_id", id.Hex()),
	)

	app := &model.Application{}
	if err := mongo.Repo.FindByID(ctx, app, id); err != nil {
		log.Error("get application failed", zap.Error(err))
		return nil, err
	}
	if app.DeletedAt != nil {
		log.Warn("application already deleted")
		return nil, mongoDriver.ErrNoDocuments
	}

	log.Debug("application fetched", zap.String("application_name", app.Name))
	return app, nil
}
