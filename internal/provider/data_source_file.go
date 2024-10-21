// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// This should be a resource. However, it's a data source to improve performance and because it doesn't matter for our intended use-case.
// This resource always creates or updates all files given with the new content.

package provider

import (
	"context"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &fileDataSource{}
)

type fileDataSource struct {
}

type fileDataSourceModel struct {
	Files           []*fileModel `tfsdk:"files"`
	AddNewlineAtEnd types.Bool   `tfsdk:"add_newline_at_end"`
}

type fileModel struct {
	Filename     types.String `tfsdk:"filename"`
	FileContents types.String `tfsdk:"file_contents"`
}

func NewFileDataSource() datasource.DataSource {
	return &fileDataSource{}
}

func (r *fileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (r *fileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"files": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"filename": schema.StringAttribute{
							Description: "Filename to create.",
							Required:    true,
						},
						"file_contents": schema.StringAttribute{
							Description: "Text to put in the file",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"add_newline_at_end": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (r *fileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan fileDataSourceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var osLineEnding string
	if runtime.GOOS == "windows" {
		osLineEnding = "\r\n"
	} else {
		osLineEnding = "\n"
	}

	for _, file := range plan.Files {
		createOrUpdateSingleFile(file, osLineEnding, plan.AddNewlineAtEnd.ValueBool(), &resp.Diagnostics)
		// Unclear if this is the best way to do this - don't save the file contents in the state
		file.FileContents = types.StringNull()
	}

	resp.State.Set(ctx, plan)
}

func createOrUpdateSingleFile(file *fileModel, osLineEnding string, addNewlineAtEnd bool, diag *diag.Diagnostics) {
	editedContents := file.FileContents.ValueString()
	if addNewlineAtEnd && !strings.HasSuffix(editedContents, osLineEnding) {
		editedContents = editedContents + osLineEnding
	}
	fileBytes := []byte(editedContents)

	// Just overwrite anything existing - likely to be faster than checking if it exists and matches the content
	err := os.WriteFile(file.Filename.ValueString(), fileBytes, 0644)
	if err != nil {
		diag.AddError("Failed to write file.", err.Error())
		return
	}
}
