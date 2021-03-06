package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *OperatorZyncOptionsProvider) GetZyncOptions() (*component.ZyncOptions, error) {
	optProv := component.ZyncOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)

	err := o.setSecretBasedOptions(&optProv)
	if err != nil {
		return nil, err
	}

	o.setResourceRequirementsOptions(&optProv)
	o.setReplicas(&optProv)

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Zync Options - %s", err)
	}
	return res, nil
}

func (o *OperatorZyncOptionsProvider) setSecretBasedOptions(zob *component.ZyncOptionsBuilder) error {
	err := o.setZyncSecretOptions(zob)
	if err != nil {
		return fmt.Errorf("unable to create Zync Secret Options - %s", err)
	}

	return nil
}

func (o *OperatorZyncOptionsProvider) setZyncSecretOptions(zob *component.ZyncOptionsBuilder) error {
	defaultZyncSecretKeyBase := oprand.String(16)
	defaultZyncDatabasePassword := oprand.String(16)
	defaultZyncAuthenticationToken := oprand.String(16)

	currSecret, err := helper.GetSecret(component.ZyncSecretName, o.Namespace, o.Client)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// If a field of a secret already exists in the deployed secret then
	// We do not modify it. Otherwise we set a default value
	secretData := currSecret.Data
	zob.SecretKeyBase(helper.GetSecretDataValueOrDefault(secretData, component.ZyncSecretKeyBaseFieldName, defaultZyncSecretKeyBase))
	zob.DatabasePassword(helper.GetSecretDataValueOrDefault(secretData, component.ZyncSecretDatabasePasswordFieldName, defaultZyncDatabasePassword))
	zob.AuthenticationToken(helper.GetSecretDataValueOrDefault(secretData, component.ZyncSecretAuthenticationTokenFieldName, defaultZyncAuthenticationToken))

	return nil
}

func (o *OperatorZyncOptionsProvider) setResourceRequirementsOptions(b *component.ZyncOptionsBuilder) {
	if !*o.APIManagerSpec.ResourceRequirementsEnabled {
		b.ContainerResourceRequirements(v1.ResourceRequirements{})
		b.QueContainerResourceRequirements(v1.ResourceRequirements{})
		b.DatabaseContainerResourceRequirements(v1.ResourceRequirements{})
	}
}

func (o *OperatorZyncOptionsProvider) setReplicas(zob *component.ZyncOptionsBuilder) {
	zob.ZyncReplicas(int32(*o.APIManagerSpec.Zync.AppSpec.Replicas))
	zob.ZyncQueReplicas(int32(*o.APIManagerSpec.Zync.QueSpec.Replicas))
}
