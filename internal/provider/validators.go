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
	for k, v := range req.ConfigValue.Elements() {
		tflog.Info(ctx, fmt.Sprintf("reqmap %s=%v", k, v.Type(ctx)))
	}
}

func (v IdentifiedWithValidator) Description(context.Context) string {
	return "ABOBA"
}

func (v IdentifiedWithValidator) MarkdownDescription(context.Context) string {
	return "ABOBA MD"
}

var clickHouseIdentifierValidator = stringvalidator.RegexMatches(
	regexp.MustCompile("^[a-zA-Z0-9_]+$"),
	"Should contain only latin letters, digits and _."+
		" Should not be empty.",
)

type grantEntityValidator struct {
}

func (v grantEntityValidator) Description(context.Context) string {
	return "Value should be a name of ClickHouse table, view, dict, etc " +
		"(regex=\"^[a-zA-Z0-9_]+$\") or \"*\""
}

func (v grantEntityValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v grantEntityValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if value == "*" {
		return
	}

	regexValidator := stringvalidator.RegexMatches(
		regexp.MustCompile("^[a-zA-Z0-9_]+$"),
		"Should contain only latin letters, digits and _."+
			" Should not be empty.",
	)
	regexValidator.ValidateString(ctx, request, response)
}
