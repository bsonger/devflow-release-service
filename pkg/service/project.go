package service

import (
	"context"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-release-service/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var ProjectService = NewProjectService()

type projectService struct{}

func NewProjectService() *projectService {
	return &projectService{}
}

func (s *projectService) Create(ctx context.Context, project *model.Project) (primitive.ObjectID, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "create_project"),
	)

	project.ApplyDefaults()
	if err := mongo.Repo.Create(ctx, project); err != nil {
		log.Error("create project failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log.Info("project created", zap.String("project_id", project.GetID().Hex()), zap.String("project_key", project.Key))
	return project.GetID(), nil
}

func (s *projectService) Get(ctx context.Context, id primitive.ObjectID) (*model.Project, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "get_project"),
		zap.String("project_id", id.Hex()),
	)

	project := &model.Project{}
	if err := mongo.Repo.FindByID(ctx, project, id); err != nil {
		log.Error("get project failed", zap.Error(err))
		return nil, err
	}
	if project.DeletedAt != nil {
		log.Warn("project already deleted")
		return nil, mongoDriver.ErrNoDocuments
	}

	log.Debug("project fetched", zap.String("project_key", project.Key))
	return project, nil
}

func (s *projectService) Update(ctx context.Context, project *model.Project) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "update_project"),
		zap.String("project_id", project.GetID().Hex()),
	)

	current := &model.Project{}
	if err := mongo.Repo.FindByID(ctx, current, project.GetID()); err != nil {
		log.Error("load project failed", zap.Error(err))
		return err
	}
	if current.DeletedAt != nil {
		log.Warn("update skipped for deleted project")
		return mongoDriver.ErrNoDocuments
	}

	project.CreatedAt = current.CreatedAt
	project.DeletedAt = current.DeletedAt
	project.WithUpdateDefault()
	project.ApplyDefaults()

	if err := mongo.Repo.Update(ctx, project); err != nil {
		log.Error("update project failed", zap.Error(err))
		return err
	}

	if current.Name != project.Name {
		if err := s.syncApplicationProjectNames(ctx, project.GetID(), project.Name); err != nil {
			log.Error("sync project name to applications failed", zap.Error(err))
			return err
		}
	}

	log.Info("project updated", zap.String("project_key", project.Key))
	return nil
}

func (s *projectService) Delete(ctx context.Context, id primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "delete_project"),
		zap.String("project_id", id.Hex()),
	)

	now := time.Now()
	update := primitive.M{
		"$set": primitive.M{
			"deleted_at": now,
			"updated_at": now,
			"status":     model.ProjectArchived,
		},
	}

	if err := mongo.Repo.UpdateByID(ctx, &model.Project{}, id, update); err != nil {
		log.Error("delete project failed", zap.Error(err))
		return err
	}

	log.Info("project deleted")
	return nil
}

func (s *projectService) List(ctx context.Context, filter primitive.M) ([]model.Project, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "list_projects"),
		zap.Any("filter", filter),
	)

	var projects []model.Project
	if err := mongo.Repo.List(ctx, &model.Project{}, filter, &projects); err != nil {
		log.Error("list projects failed", zap.Error(err))
		return nil, err
	}

	log.Debug("projects listed", zap.Int("count", len(projects)))
	return projects, nil
}

func (s *projectService) ListApplications(ctx context.Context, projectID primitive.ObjectID) ([]model.Application, error) {
	project, err := s.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}

	filter := primitive.M{
		"deleted_at": primitive.M{"$exists": false},
		"$or": []primitive.M{
			{"project_id": projectID},
			{
				"project_name": project.Name,
				"$or": []primitive.M{
					{"project_id": primitive.M{"$exists": false}},
					{"project_id": nil},
				},
			},
		},
	}

	return ApplicationService.List(ctx, filter)
}

func (s *projectService) syncApplicationProjectNames(ctx context.Context, projectID primitive.ObjectID, projectName string) error {
	return mongo.Repo.UpdateMany(ctx, &model.Application{}, bson.M{"project_id": projectID}, bson.M{
		"$set": bson.M{
			"project_name": projectName,
			"updated_at":   time.Now(),
		},
	})
}
