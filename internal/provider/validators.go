package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"
)

type IdentifiedWithValidator struct{}

func (v IdentifiedWithValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	//tflog.Info(ctx, fmt.Sprintf("req config: %v", req.Config))
	//tflog.Info(ctx, fmt.Sprintf("req config value: %v", req.ConfigValue))
	//tflog.Info(ctx, fmt.Sprintf("req path: %v", req.Path))
	//tflog.Info(ctx, fmt.Sprintf("req path expr: %v", req.PathExpression))

	//tflog.Info(ctx, fmt.Sprintf("req map: %v", req.ConfigValue.Elements()))

	for k, v := range req.ConfigValue.Elements() {
		tflog.Info(ctx, fmt.Sprintf("reqmap %s=%v", k, v.Type(ctx)))
	}

	//tflog.Info(ctx, fmt.Sprintf("res: %v", resp))
}

func (v IdentifiedWithValidator) Description(context.Context) string {
	return "ABOBA"
}

func (v IdentifiedWithValidator) MarkdownDescription(context.Context) string {
	return "ABOBA MD"
}

var ClickHouseIdentifierValidator = stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9_]+$"), "Should contain only latin letters, digits and _")
