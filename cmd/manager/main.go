package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/apis"
	"github.com/3scale/3scale-operator/pkg/controller"
	"github.com/3scale/3scale-operator/version"
	"github.com/prometheus/client_golang/prometheus"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "3scale-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	register3scaleVersionInfoMetric()

	// Create Service object to expose the metrics port.
	metricService, err := metrics.ExposeMetricsPort(ctx, metricsPort)
	if err != nil {
		log.Error(err, "")
	}

	err = updateMetricServiceLabels(ctx, cfg, metricService)
	if err != nil {
		log.Error(err, "Could not add monitoring-key label to operator metrics Service")
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator.
	_, err = metrics.CreateServiceMonitors(cfg, namespace, []*v1.Service{metricService})
	if err != nil {
		log.Info("Could not create ServiceMonitor object", "error", err.Error())
		// Available in operator-sdk 0.10.0
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		//if err == metrics.ErrServiceMonitorNotPresent {
		//	log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects", "error", err.Error())
		//}
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func register3scaleVersionInfoMetric() {
	threeScaleVersionInfo := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "threescale_version_info",
			Help: "3scale Operator version and product version",
			ConstLabels: prometheus.Labels{
				"operator_version": version.Version,
				"version":          product.ThreescaleRelease,
			},
		},
	)
	// Register custom metrics with the global prometheus registry
	crmetrics.Registry.MustRegister(threeScaleVersionInfo)
}

func updateMetricServiceLabels(ctx context.Context, cfg *rest.Config, service *v1.Service) error {
	if service == nil {
		return fmt.Errorf("service doesn't exist")
	}

	kclient, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}

	updatedLabels := map[string]string{
		"monitoring-key": "middleware",
		"app":            appsv1alpha1.Default3scaleAppLabel,
	}
	for k, v := range service.ObjectMeta.Labels {
		updatedLabels[k] = v
	}
	service.ObjectMeta.Labels = updatedLabels

	err = kclient.Update(ctx, service)
	if err != nil {
		return err
	}

	return nil
}
