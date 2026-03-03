// Package controller provides custom controller setup for the Cloudflare provider.
package controller

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// newObserveShortCircuitConnecter wraps an ExternalConnecter so that Observe
// returns ResourceExists: false immediately when the managed resource has no
// external-name annotation. This prevents Upjet's workspace from running
// "terraform refresh" with an empty resource ID, which fails for Terraform
// providers (like Cloudflare v5) whose Read function does not gracefully
// handle empty IDs.
func newObserveShortCircuitConnecter(wrapped managed.ExternalConnecter) managed.ExternalConnecter {
	return &observeShortCircuitConnecter{wrapped: wrapped}
}

type observeShortCircuitConnecter struct {
	wrapped managed.ExternalConnecter
}

func (c *observeShortCircuitConnecter) Connect(ctx context.Context, mg xpresource.Managed) (managed.ExternalClient, error) {
	ext, err := c.wrapped.Connect(ctx, mg)
	if err != nil {
		return nil, err
	}
	return &observeShortCircuitClient{wrapped: ext}, nil
}

type observeShortCircuitClient struct {
	wrapped managed.ExternalClient
}

func (e *observeShortCircuitClient) Observe(ctx context.Context, mg xpresource.Managed) (managed.ExternalObservation, error) {
	// If no external-name annotation, the resource hasn't been created yet.
	// Skip terraform refresh (which fails with empty ID for Cloudflare v5)
	// and tell the reconciler to proceed to Create.
	if meta.GetExternalName(mg) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	return e.wrapped.Observe(ctx, mg)
}

func (e *observeShortCircuitClient) Create(ctx context.Context, mg xpresource.Managed) (managed.ExternalCreation, error) {
	return e.wrapped.Create(ctx, mg)
}

func (e *observeShortCircuitClient) Update(ctx context.Context, mg xpresource.Managed) (managed.ExternalUpdate, error) {
	return e.wrapped.Update(ctx, mg)
}

func (e *observeShortCircuitClient) Delete(ctx context.Context, mg xpresource.Managed) (managed.ExternalDelete, error) {
	return e.wrapped.Delete(ctx, mg)
}

func (e *observeShortCircuitClient) Disconnect(ctx context.Context) error {
	type disconnecter interface {
		Disconnect(context.Context) error
	}
	if d, ok := e.wrapped.(disconnecter); ok {
		return d.Disconnect(ctx)
	}
	return nil
}
