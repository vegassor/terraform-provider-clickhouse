package provider

import (
	"context"
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

	if tableModel.PartitionBy == "" {
		// PlanValue should already be empty list due to Default function
		return
	}

	var partitionByParams []string
	re := regexp.MustCompile(`\w+\((\w+)\)`)
	matches := re.FindStringSubmatch(tableModel.PartitionBy)
	if len(matches) > 1 {
		partitionByParams = matches[1:]
	} else {
		partitionByParams = []string{tableModel.PartitionBy}
	}

	val, ds := basetypes.NewListValueFrom(ctx, types.StringType, partitionByParams)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = val
}
