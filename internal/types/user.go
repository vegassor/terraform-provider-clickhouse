package types

import "github.com/hashicorp/terraform-plugin-framework/types"

type sha256Hash struct {
	Hash string  `tfsdk:"hash"`
	Salt *string `tfsdk:"salt"`
}

type sha256Password struct {
	Password string `tfsdk:"hash"`
}

type identifiedWith struct {
	Sha256Hash     *sha256Hash     `tfsdk:"sha256_hash"`
	Sha256Password *sha256Password `tfsdk:"sha256_password"`
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	Name           string         `tfsdk:"name"`
	IdentifiedWith identifiedWith `tfsdk:"identified_with"`
	Hosts          types.List     `tfsdk:"hosts"`
}
