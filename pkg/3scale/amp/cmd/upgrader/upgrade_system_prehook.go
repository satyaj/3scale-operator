package upgrader

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpgradeSystemPreHook(cl client.Client, ns string) error {
	existing := &appsv1.DeploymentConfig{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: ns}, existing)
	if err != nil {
		return err
	}

	system, err := GetSystemComponent()
	if err != nil {
		return err
	}

	desiredDeploymentConfig := system.AppDeploymentConfig()
	changed := ensureDeploymentConfigPreHookPodEnvVars(desiredDeploymentConfig, existing)
	tmpChanged := ensureDeploymentConfigPreHookPodCommand(desiredDeploymentConfig, existing)
	changed = changed || tmpChanged

	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}

func ensureDeploymentConfigPreHookPodEnvVars(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false
	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPreHookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod

	// replace SMTP_* vars
	for i := range existingPreHookPod.Env {
		if existingPreHookPod.Env[i].Name == "SMTP_ADDRESS" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_ADDRESS")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}

		if existingPreHookPod.Env[i].Name == "SMTP_USER_NAME" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_USER_NAME")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}

		if existingPreHookPod.Env[i].Name == "SMTP_PASSWORD" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_PASSWORD")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}

		if existingPreHookPod.Env[i].Name == "SMTP_DOMAIN" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_DOMAIN")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}

		if existingPreHookPod.Env[i].Name == "SMTP_PORT" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_PORT")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}

		if existingPreHookPod.Env[i].Name == "SMTP_AUTHENTICATION" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_AUTHENTICATION")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}

		if existingPreHookPod.Env[i].Name == "SMTP_OPENSSL_VERIFY_MODE" {
			desiredEnvVar := findEnvByNameOrPanic(desiredPreHookPod.Env, "SMTP_OPENSSL_VERIFY_MODE")
			if !reflect.DeepEqual(existingPreHookPod.Env[i].ValueFrom, desiredEnvVar.ValueFrom) {
				existingPreHookPod.Env[i].ValueFrom = desiredEnvVar.ValueFrom
				changed = true
			}
		}
	}

	if !findEnvByName(existingPreHookPod.Env, component.SystemSecretSystemSeedMasterAccessTokenFieldName) {
		// Add MASTER_ACCESS_TOKEN ref
		existingPreHookPod.Env = append(existingPreHookPod.Env,
			v1.EnvVar{
				Name: component.SystemSecretSystemSeedMasterAccessTokenFieldName,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: component.SystemSecretSystemSeedSecretName,
						},
						Key: component.SystemSecretSystemSeedMasterAccessTokenFieldName,
					},
				},
			})
		changed = true
	}

	return changed
}

func ensureDeploymentConfigPreHookPodCommand(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false
	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPrehookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod
	if !reflect.DeepEqual(existingPrehookPod.Command, desiredPreHookPod.Command) {
		existingPrehookPod.Command = desiredPreHookPod.Command
		changed = true
	}
	return changed
}

func findEnvByNameOrPanic(a []v1.EnvVar, x string) v1.EnvVar {
	for _, n := range a {
		if x == n.Name {
			return n
		}
	}

	panic(fmt.Sprintf("env var %s not found", x))
}

func findEnvByName(a []v1.EnvVar, x string) bool {
	for _, n := range a {
		if x == n.Name {
			return true
		}
	}
	return false
}
