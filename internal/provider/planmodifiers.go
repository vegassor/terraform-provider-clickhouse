package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"regexp"
)

type partitionByPlanModifier struct{}

func (m partitionByPlanModifier) Description(_ context.Context) string {
	return "sets value equal to partition_by if value is not configured"
}

func (m partitionByPlanModifier) MarkdownDescription(_ context.Context) string {
	return "sets value equal to `partition_by` if value is not configured"
}

func (m partitionByPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	if (req.PlanValue.IsUnknown() || req.PlanValue.IsNull()) && !req.StateValue.IsUnknown() {
		resp.RequiresReplace = true
		return
	}

	var tableModel TableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tableModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	partitonBy := tableModel.PartitionBy.ValueString()
	if partitonBy == "" {
		// PlanValue should already be empty list due to Default function
		return
	}

	var partitionByParams []string
	re := regexp.MustCompile(`\w+\((\w+)\)`)
	matches := re.FindStringSubmatch(partitonBy)
	if len(matches) > 1 {
		partitionByParams = matches[1:]
	} else {
		partitionByParams = []string{partitonBy}
	}

	val, ds := basetypes.NewListValueFrom(ctx, types.StringType, partitionByParams)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = val
}

type fullNamePlanModifier struct{}

func (m fullNamePlanModifier) Description(_ context.Context) string {
	return "Constructs full_name from database and table name."
}

func (m fullNamePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Constructs `full_name` from `database` and table's `name`."
}

func (m fullNamePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	var db, name string
	resp.Diagnostics.Append(
		req.Plan.GetAttribute(ctx, path.Root("database"), &db)...,
	)
	resp.Diagnostics.Append(
		req.Plan.GetAttribute(ctx, path.Root("name"), &name)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	if db == "" || name == "" {
		return
	}

	resp.PlanValue = types.StringValue(db + "." + name)
}
