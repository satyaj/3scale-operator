package upgrader

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpgradeImages(cl client.Client, ns string) error {
	ampImages, err := GetAMPImagesComponent()
	if err != nil {
		return err
	}

	imagestreams := []*imagev1.ImageStream{
		ampImages.SystemImageStream(),
		ampImages.APICastImageStream(),
		ampImages.BackendImageStream(),
		ampImages.ZyncImageStream(),
		ampImages.SystemMemcachedImageStream(),
		ampImages.ZyncDatabasePostgreSQLImageStream(),
	}
	reconciler := operator.NewImageStreamGenericReconciler()

	for _, desired := range imagestreams {
		existing := &imagev1.ImageStream{}
		err := cl.Get(
			context.TODO(),
			types.NamespacedName{Name: desired.Name, Namespace: ns},
			existing)
		if err != nil {
			return err
		}

		if reconciler.IsUpdateNeeded(desired, existing) {
			err := cl.Update(context.TODO(), existing)
			if err != nil {
				return err
			}
			fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		}
	}

	optionalImagestreams := []*imagev1.ImageStream{
		ampImages.SystemImageStream(),
		ampImages.APICastImageStream(),
		ampImages.BackendImageStream(),
		ampImages.ZyncImageStream(),
		ampImages.SystemMemcachedImageStream(),
		ampImages.ZyncDatabasePostgreSQLImageStream(),
	}

	for _, desired := range optionalImagestreams {
		existing := &imagev1.ImageStream{}
		err := cl.Get(
			context.TODO(),
			types.NamespacedName{Name: desired.Name, Namespace: ns},
			existing)
		if err != nil {
			return err
		}

		if reconciler.IsUpdateNeeded(desired, existing) {
			err := cl.Update(context.TODO(), existing)
			if err != nil {
				return err
			}
			fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		}
	}
	return nil
}

func GetAMPImagesComponent() (*component.AmpImages, error) {
	optProv := component.AmpImagesOptionsBuilder{}
	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AMPRelease(product.ThreescaleRelease)
	optProv.ApicastImage(component.ApicastImageURL())
	optProv.BackendImage(component.BackendImageURL())
	optProv.SystemImage(component.SystemImageURL())
	optProv.ZyncImage(component.ZyncImageURL())
	optProv.ZyncDatabasePostgreSQLImage(component.ZyncPostgreSQLImageURL())
	optProv.SystemMemcachedImage(component.SystemMemcachedImageURL())
	optProv.InsecureImportPolicy(v1alpha1.DefaultImageStreamImportInsecure)

	otions, err := optProv.Build()
	if err != nil {
		return nil, err
	}
	return component.NewAmpImages(otions), nil
}
