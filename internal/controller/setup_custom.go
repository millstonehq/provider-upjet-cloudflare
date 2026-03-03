package controller

import (
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	xpfeature "github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	tjcontroller "github.com/crossplane/upjet/v2/pkg/controller"
	"github.com/crossplane/upjet/v2/pkg/controller/handler"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	dnsv1alpha1 "github.com/millstonehq/provider-upjet-cloudflare/apis/dns/v1alpha1"
	zonev1alpha1 "github.com/millstonehq/provider-upjet-cloudflare/apis/zone/v1alpha1"
	"github.com/millstonehq/provider-upjet-cloudflare/internal/controller/providerconfig"
	"github.com/millstonehq/provider-upjet-cloudflare/internal/features"
)

// SetupCustom creates all controllers with the observe short-circuit wrapper
// applied to resource controllers. This prevents terraform refresh from failing
// on resources that haven't been created yet (Cloudflare v5 Read fails with
// empty IDs).
func SetupCustom(mgr ctrl.Manager, o tjcontroller.Options) error {
	// ProviderConfig uses the standard setup (no terraform workspace)
	if err := providerconfig.Setup(mgr, o); err != nil {
		return err
	}
	// DNS Record with observe short-circuit
	if err := setupResourceWithWrapper(mgr, o, "cloudflare_dns_record", dnsv1alpha1.Record_GroupVersionKind, &dnsv1alpha1.Record{}, &dnsv1alpha1.RecordList{}); err != nil {
		return err
	}
	// Zone with observe short-circuit
	if err := setupResourceWithWrapper(mgr, o, "cloudflare_zone", zonev1alpha1.Zone_GroupVersionKind, &zonev1alpha1.Zone{}, &zonev1alpha1.ZoneList{}); err != nil {
		return err
	}
	return nil
}

type managedListObject interface {
	xpresource.ManagedList
}

func setupResourceWithWrapper(
	mgr ctrl.Manager,
	o tjcontroller.Options,
	terraformResourceName string,
	gvk schema.GroupVersionKind,
	managedObj xpresource.Managed,
	managedList managedListObject,
) error {
	name := managed.ControllerName(gvk.String())
	var initializers managed.InitializerChain
	eventHandler := handler.NewEventHandler(handler.WithLogger(o.Logger.WithValues("gvk", gvk)))

	// Wrap the standard connector with the observe short-circuit
	connector := newObserveShortCircuitConnecter(
		tjcontroller.NewConnector(mgr.GetClient(), o.WorkspaceStore, o.SetupFn,
			o.Provider.Resources[terraformResourceName],
			tjcontroller.WithLogger(o.Logger),
			tjcontroller.WithConnectorEventHandler(eventHandler),
		),
	)

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(connector),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithFinalizer(terraform.NewWorkspaceFinalizer(o.WorkspaceStore, xpresource.NewAPIFinalizer(mgr.GetClient(), managed.FinalizerName))),
		managed.WithTimeout(3 * time.Minute),
		managed.WithInitializers(initializers),
		managed.WithPollInterval(o.PollInterval),
	}
	if o.PollJitter != 0 {
		opts = append(opts, managed.WithPollJitterHook(o.PollJitter))
	}
	if o.Features.Enabled(features.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}
	if o.MetricOptions != nil {
		opts = append(opts, managed.WithMetricRecorder(o.MetricOptions.MRMetrics))
	}
	if o.Features.Enabled(xpfeature.EnableAlphaChangeLogs) {
		opts = append(opts, managed.WithChangeLogger(o.ChangeLogOptions.ChangeLogger))
	}
	if o.MetricOptions != nil && o.MetricOptions.MRStateMetrics != nil {
		stateMetricsRecorder := statemetrics.NewMRStateRecorder(
			mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, managedList, o.MetricOptions.PollStateMetricInterval,
		)
		if err := mgr.Add(stateMetricsRecorder); err != nil {
			return err
		}
	}

	r := managed.NewReconciler(mgr, xpresource.ManagedKind(gvk), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(xpresource.DesiredStateChanged()).
		Watches(managedObj, eventHandler).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}
