package argoclient

import (
	"context"
	"fmt"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	argoapi "github.com/argoproj/argo-cd/v3/pkg/client/clientset/versioned"
	"github.com/bsonger/devflow-service-common/loggingx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

var Client *argoapi.Clientset

const namespace = "argocd"

type applicationAPI interface {
	Create(ctx context.Context, app *appv1.Application, opts metav1.CreateOptions) (*appv1.Application, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*appv1.Application, error)
	Update(ctx context.Context, app *appv1.Application, opts metav1.UpdateOptions) (*appv1.Application, error)
}

func Init(config *rest.Config) error {
	var err error
	Client, err = argoapi.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create argo cd client: %w", err)
	}
	loggingx.Logger.Info("argo cd client initialized")
	return nil
}

func CreateApplication(ctx context.Context, app *appv1.Application) error {
	_, err := Client.ArgoprojV1alpha1().Applications(namespace).Create(ctx, app, metav1.CreateOptions{})
	return err
}

func UpdateApplication(ctx context.Context, app *appv1.Application) error {
	return applyApplication(ctx, Client.ArgoprojV1alpha1().Applications(namespace), app)
}

func applyApplication(ctx context.Context, applications applicationAPI, app *appv1.Application) error {
	current, err := applications.Get(ctx, app.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err = applications.Create(ctx, app, metav1.CreateOptions{})
			return err
		}
		return err
	}

	current.Spec = app.Spec
	current.Annotations = app.Annotations
	current.Labels = app.Labels

	_, err = applications.Update(ctx, current, metav1.UpdateOptions{})
	return err
}
