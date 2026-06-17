// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var (
	_ resource.Resource                = &certificateResource{}
	_ resource.ResourceWithImportState = &certificateResource{}
)

var certificateReqOpts = opnsense.ReqOpts{
	AddEndpoint:         "/api/acmeclient/certificates/add",
	GetEndpoint:         "/api/acmeclient/certificates/get",
	UpdateEndpoint:      "/api/acmeclient/certificates/set",
	DeleteEndpoint:      "/api/acmeclient/certificates/del",
	SearchEndpoint:      "/api/acmeclient/certificates/search",
	ReconfigureEndpoint: "/api/acmeclient/service/reconfigure",
	Monad:               "certificate",
}

const certificateSignEndpoint = "/api/acmeclient/certificates/sign"

type certificateResource struct{ client *opnsense.Client }

func newCertificateResource() resource.Resource { return &certificateResource{} }

func (r *certificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_certificate"
}

func (r *certificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	r.client = client
}

func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, interval, err := plan.issuanceWaitConfig()
	if err != nil {
		resp.Diagnostics.AddError("Invalid ACME issuance wait configuration", err.Error())
		return
	}
	uuid, err := opnsense.Add(ctx, r.client, certificateReqOpts, plan.toAPI(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Error creating ACME certificate", fmt.Sprintf("%s", err))
		return
	}
	result, err := signAndWaitForCertificateIssuance(ctx, r.client, uuid, timeout, interval)
	if err != nil {
		if deleteErr := opnsense.Delete(ctx, r.client, certificateReqOpts, uuid); deleteErr != nil {
			err = fmt.Errorf("%w; additionally failed to clean up created certificate %s: %s", err, uuid, deleteErr)
		}
		resp.Diagnostics.AddError("Error issuing ACME certificate", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, uuid)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := opnsense.Get[certificateAPIResponse](ctx, r.client, certificateReqOpts, state.ID.ValueString())
	if err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading ACME certificate", fmt.Sprintf("%s", err))
		return
	}
	state.fromAPI(ctx, result, state.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *certificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	timeout, interval, err := plan.issuanceWaitConfig()
	if err != nil {
		resp.Diagnostics.AddError("Invalid ACME issuance wait configuration", err.Error())
		return
	}
	requiresRemoteUpdate := plan.requiresRemoteUpdate(state)
	requiresIssuance := plan.requiresIssuance(state)
	if requiresRemoteUpdate {
		if err := opnsense.Update(ctx, r.client, certificateReqOpts, plan.toAPI(ctx), id); err != nil {
			resp.Diagnostics.AddError("Error updating ACME certificate", fmt.Sprintf("%s", err))
			return
		}
	}
	if !requiresIssuance {
		result, err := opnsense.Get[certificateAPIResponse](ctx, r.client, certificateReqOpts, id)
		if err != nil {
			resp.Diagnostics.AddError("Error reading ACME certificate after update", fmt.Sprintf("%s", err))
			return
		}
		plan.fromAPI(ctx, result, id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	result, err := signAndWaitForCertificateIssuance(ctx, r.client, id, timeout, interval)
	if err != nil {
		resp.Diagnostics.AddError("Error issuing ACME certificate after update", fmt.Sprintf("%s", err))
		return
	}
	plan.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := opnsense.Delete(ctx, r.client, certificateReqOpts, state.ID.ValueString()); err != nil {
		var nf *opnsense.NotFoundError
		if errors.As(err, &nf) {
			return
		}
		resp.Diagnostics.AddError("Error deleting ACME certificate", fmt.Sprintf("%s", err))
	}
}

func (r *certificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func signAndWaitForCertificateIssuance(ctx context.Context, client *opnsense.Client, uuid string, timeout, interval time.Duration) (*certificateAPIResponse, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("issuance timeout must be positive")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("issuance poll interval must be positive")
	}

	if err := signCertificate(ctx, client, uuid); err != nil {
		return nil, err
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		issued, row, err := certificateIssued(waitCtx, client, uuid)
		if err != nil {
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return nil, certificateIssuanceTimeoutError(uuid, timeout)
			}
			return nil, err
		}
		if issued {
			result, err := opnsense.Get[certificateAPIResponse](waitCtx, client, certificateReqOpts, uuid)
			if err != nil {
				return nil, fmt.Errorf("read issued ACME certificate %s: %w", uuid, err)
			}
			if result.CertRefID == "" {
				result.CertRefID = row.CertRefID
			}
			if result.StatusCode == "" {
				result.StatusCode = row.StatusCode
			}
			if result.Status == "" {
				result.Status = row.Status
			}
			return result, nil
		}

		timer := time.NewTimer(interval)
		select {
		case <-waitCtx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			if !errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return nil, waitCtx.Err()
			}
			return nil, certificateIssuanceTimeoutError(uuid, timeout)
		case <-timer.C:
		}
	}
}

func certificateIssuanceTimeoutError(uuid string, timeout time.Duration) error {
	return fmt.Errorf("ACME certificate %s issuance timed out after %s waiting for statusCode 200 and non-empty certRefId", uuid, timeout)
}

func signCertificate(ctx context.Context, client *opnsense.Client, uuid string) error {
	if err := client.LockMutex(ctx); err != nil {
		return fmt.Errorf("sign %s: %w", certificateSignEndpoint, err)
	}
	defer client.UnlockMutex()

	endpoint := certificateSignEndpoint + "/" + url.PathEscape(uuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.BaseURL()+endpoint, nil)
	if err != nil {
		return fmt.Errorf("sign %s: %w", endpoint, err)
	}

	response, err := client.HTTPClient().Do(req) //nolint:gosec // URL from provider-configured client plus fixed endpoint.
	if err != nil {
		return opnsense.NewServerError(endpoint, err)
	}
	defer func() { _ = response.Body.Close() }()

	if httpErr := opnsense.CheckHTTPError(response.StatusCode, endpoint); httpErr != nil {
		return httpErr
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("sign %s: failed to read response: %w", endpoint, err)
	}
	if err := parseCertificateSignResponse(body); err != nil {
		return fmt.Errorf("sign %s: %w", endpoint, err)
	}
	return nil
}

func certificateIssued(ctx context.Context, client *opnsense.Client, uuid string) (bool, certificateSearchRow, error) {
	rows, err := opnsense.Search[certificateSearchRow](ctx, client, certificateReqOpts, opnsense.SearchParams{})
	if err != nil {
		return false, certificateSearchRow{}, err
	}

	for _, row := range rows {
		if row.UUID != uuid {
			continue
		}
		if strings.TrimSpace(row.StatusCode.String()) == "200" && strings.TrimSpace(row.CertRefID) != "" {
			return true, row, nil
		}
		return false, row, nil
	}

	return false, certificateSearchRow{}, nil
}
