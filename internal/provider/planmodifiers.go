package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"regexp"
	"strings"
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

type CompositeNamePlanModifier struct {
	paths     []path.Path
	separator string
}

func NewCompositePlanModifier(p []path.Path, sep string) CompositeNamePlanModifier {
	return CompositeNamePlanModifier{paths: p, separator: sep}
}

func NewCompositePlanModifierFromStr(p []string, sep string) CompositeNamePlanModifier {
	paths := make([]path.Path, 0, len(p))

	for _, s := range p {
		paths = append(paths, path.Root(s))
	}

	return CompositeNamePlanModifier{paths: paths, separator: sep}
}

func (m CompositeNamePlanModifier) Description(_ context.Context) string {
	return "Takes multiple values from config and combines them into a single string using a separator."
}

func (m CompositeNamePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Takes multiple values from config and combines them into a single string using a separator."
}

func (m CompositeNamePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	var values []string
	for _, p := range m.paths {
		var value string
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, p, &value)...)
		values = append(values, value)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = types.StringValue(strings.Join(values, m.separator))
}
