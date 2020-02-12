package upgrader

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func RegisterOpenShiftAPIGroups(s *runtime.Scheme) error {
	var addToSchemes runtime.SchemeBuilder
	addToSchemes = append(addToSchemes,
		appsv1.Install,
		imagev1.Install,
		routev1.Install,
	)

	return addToSchemes.AddToScheme(s)
}

func GetSystemComponent() (*component.System, error) {
	optProv := component.SystemOptionsBuilder{}

	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AmpRelease("-")
	optProv.ApicastRegistryURL("-")
	optProv.TenantName("-")
	optProv.WildcardDomain("-")
	optProv.AdminAccessToken("-")
	optProv.AdminPassword("-")
	optProv.AdminUsername("-")
	optProv.ApicastAccessToken("-")
	optProv.MasterAccessToken("-")
	optProv.MasterUsername("-")
	optProv.MasterPassword("-")
	optProv.AppSecretKeyBase("-")
	optProv.BackendSharedSecret("-")
	optProv.MasterName("-")
	systemOptions, err := optProv.Build()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(systemOptions), nil
}
