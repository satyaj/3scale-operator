package upgrader

import (
	"context"
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	"reflect"

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
			existingPreHookPod.Env[i].ValueFrom = findEnvByName(desiredPreHookPod.Env, "SMPT_ADDRESS").ValueFrom
			changed = true
		}

		if existingPreHookPod.Env[i].Name == "SMTP_USER_NAME" {
			existingPreHookPod.Env[i].ValueFrom = findEnvByName(desiredPreHookPod.Env, "SMTP_USER_NAME").ValueFrom
			changed = true
		}

		if existingPreHookPod.Env[i].Name == "SMTP_PASSWORD" {
			existingPreHookPod.Env[i].ValueFrom = findEnvByName(desiredPreHookPod.Env, "SMTP_PASSWORD").ValueFrom
			changed = true
		}
		...
	}

	// Add MASTER_TOKEN ref
	todo
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

func findEnvByName(a []corev1.EnvVar, x string) corev1.EnvVar {
	for i, n := range a {
		if x == n.Name {
			return n
		}
	}

	panic(fmt.Sprintf("env var %s not found", x))
}
