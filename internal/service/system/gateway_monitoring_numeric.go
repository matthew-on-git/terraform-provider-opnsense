// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Bounds follow OPNsense 26.7.8 Routing/Gateways.xml at core commit
// 96978001a8343669154e4492172ce0a07dac7e4b.
// Optional IntegerField values use strings so empty/default remains distinct
// from zero. This also allows an explicit empty update to clear a saved override.
type gatewayIntegerValidator struct{ min, max int64 }

func (v gatewayIntegerValidator) Description(context.Context) string {
	return "must be empty or a decimal integer within the appliance field bounds"
}

func (v gatewayIntegerValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v gatewayIntegerValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := gatewayIntegerValue(req.ConfigValue.ValueString(), v.min, v.max); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Gateway Monitoring Value", err.Error())
	}
}

func gatewayIntegerValue(value string, minimum, maximum int64) error {
	if value == "" {
		return nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	digits := true
	for _, char := range value {
		if char < '0' || char > '9' {
			digits = false
		}
	}
	if err != nil || !digits || strconv.FormatInt(n, 10) != value || n < minimum || (maximum >= 0 && n > maximum) {
		if maximum >= 0 {
			return fmt.Errorf("expected empty or a canonical decimal integer between %d and %d", minimum, maximum)
		}
		return fmt.Errorf("expected empty or a canonical decimal integer between %d and 9223372036854775807", minimum)
	}
	return nil
}

func (api *gatewayAPIResponse) UnmarshalJSON(raw []byte) error {
	type wire gatewayAPIResponse
	var parsed wire
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	if parsed.LatencyLow != nil {
		if err := gatewayIntegerValue(*parsed.LatencyLow, 1, -1); err != nil {
			return fmt.Errorf("invalid gateway latencylow: %w", err)
		}
	}
	if parsed.LatencyHigh != nil {
		if err := gatewayIntegerValue(*parsed.LatencyHigh, 1, -1); err != nil {
			return fmt.Errorf("invalid gateway latencyhigh: %w", err)
		}
	}
	if parsed.LossLow != nil {
		if err := gatewayIntegerValue(*parsed.LossLow, 1, 99); err != nil {
			return fmt.Errorf("invalid gateway losslow: %w", err)
		}
	}
	if parsed.LossHigh != nil {
		if err := gatewayIntegerValue(*parsed.LossHigh, 1, 100); err != nil {
			return fmt.Errorf("invalid gateway losshigh: %w", err)
		}
	}
	if parsed.Interval != nil {
		if err := gatewayIntegerValue(*parsed.Interval, 1, -1); err != nil {
			return fmt.Errorf("invalid gateway interval: %w", err)
		}
	}
	if parsed.TimePeriod != nil {
		if err := gatewayIntegerValue(*parsed.TimePeriod, 1, -1); err != nil {
			return fmt.Errorf("invalid gateway time_period: %w", err)
		}
	}
	if parsed.LossInterval != nil {
		if err := gatewayIntegerValue(*parsed.LossInterval, 1, -1); err != nil {
			return fmt.Errorf("invalid gateway loss_interval: %w", err)
		}
	}
	if parsed.DataLength != nil {
		if err := gatewayIntegerValue(*parsed.DataLength, 0, -1); err != nil {
			return fmt.Errorf("invalid gateway data_length: %w", err)
		}
	}

	*api = gatewayAPIResponse(parsed)
	return nil
}

func (r *gatewayResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{gatewayThresholdValidator{}}
}

type gatewayThresholdValidator struct{}

func (gatewayThresholdValidator) Description(context.Context) string {
	return "low latency/loss thresholds must be lower than high thresholds"
}

func (v gatewayThresholdValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (gatewayThresholdValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var model GatewayResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !model.LatencyLow.IsNull() && !model.LatencyLow.IsUnknown() && !model.LatencyHigh.IsNull() && !model.LatencyHigh.IsUnknown() && model.LatencyLow.ValueString() != "" && model.LatencyHigh.ValueString() != "" {
		low, lowErr := strconv.ParseInt(model.LatencyLow.ValueString(), 10, 64)
		high, highErr := strconv.ParseInt(model.LatencyHigh.ValueString(), 10, 64)
		if lowErr == nil && highErr == nil && low >= high {
			resp.Diagnostics.AddAttributeError(path.Root("latency_low"), "Invalid Gateway Threshold Order", "The low threshold must be lower than the high threshold.")
		}
	}
	if !model.LossLow.IsNull() && !model.LossLow.IsUnknown() && !model.LossHigh.IsNull() && !model.LossHigh.IsUnknown() && model.LossLow.ValueString() != "" && model.LossHigh.ValueString() != "" {
		low, lowErr := strconv.ParseInt(model.LossLow.ValueString(), 10, 64)
		high, highErr := strconv.ParseInt(model.LossHigh.ValueString(), 10, 64)
		if lowErr == nil && highErr == nil && low >= high {
			resp.Diagnostics.AddAttributeError(path.Root("loss_low"), "Invalid Gateway Threshold Order", "The low threshold must be lower than the high threshold.")
		}
	}
}
