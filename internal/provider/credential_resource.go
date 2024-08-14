package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &credentialResource{}
	_ resource.ResourceWithConfigure = &credentialResource{}
)

func NewCredentialResource() resource.Resource {
	return &credentialResource{}
}

type credentialResource struct {
	client *sifflet.Client
}

// Metadata returns the resource type name.
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func CredentialResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		// TODO docs
		Description: "A credential resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the credential",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the credential",
				Optional:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the credential",
				Sensitive:   true,
			},
		},
	}

}

type CredentialDto struct {
	Name        string  `tfsdk:"name"`
	Description *string `tfsdk:"description"`
	Value       string  `tfsdk:"value"`
}

func (r *credentialResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CredentialResourceSchema(ctx)
}

func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan CredentialDto
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialDto := sifflet.PublicCredentialCreateDto{
		Name:        plan.Name,
		Description: plan.Description,
		Value:       plan.Value,
	}

	credentialResponse, err := r.client.PublicCreateCredential(ctx, credentialDto)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create credential",
			err.Error(),
		)
		return
	}
	// if credentialResponse.StatusCode != http.StatusCreated {

	resBody, _ := io.ReadAll(credentialResponse.Body)
	// FIXME debug logs
	tflog.Debug(ctx, "Response:  "+string(resBody))

	// if credentialResponse.StatusCode != http.StatusCreated {
	// 	if err := json.Unmarshal(resBody, &message); err != nil { // Parse []byte to go struct pointer
	// 		resp.Diagnostics.AddError(
	// 			"Can not unmarshal JSON",
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	resp.Diagnostics.AddError(
	// 		message.Title,
	// 		message.Detail,
	// 	)
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }
	//
	// var result sifflet.PublicGetCredentialResponse
	// if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
	// 	resp.Diagnostics.AddError(
	// 		"Can not unmarshal JSON",
	// 		err.Error(),
	// 	)
	// 	return
	// }
	//
	// // Map response body to schema and populate Computed attribute values
	//
	// Type := tag_struct.TagDtoType(result.Type)
	// id := result.Id.String()
	//
	// plan.CreatedBy = types.StringValue(*result.CreatedBy)
	// plan.CreatedDate = types.StringValue(strconv.FormatInt(*result.CreatedDate, 10))
	// plan.Description = result.Description
	// plan.Editable = types.BoolValue(*result.Editable)
	// plan.Id = types.StringValue(id)
	// plan.LastModifiedDate = types.StringValue(strconv.FormatInt(*result.LastModifiedDate, 10))
	// plan.ModifiedBy = types.StringValue(*result.ModifiedBy)
	// plan.Name = &result.Name
	// plan.Type = &Type
	//
	// // Set state to fully populated data
	// diags = resp.State.Set(ctx, plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
}

// Read refreshes the Terraform state with the latest data.
func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tag_struct.TagDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.String()

	itemResponse, err := r.client.GetTagById(ctx, uuid.MustParse(id))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Item",
			err.Error(),
		)
		return
	}

	resBody, _ := io.ReadAll(itemResponse.Body)
	tflog.Debug(ctx, fmt.Sprintf("Response: %d ", itemResponse.Body))

	if itemResponse.StatusCode == http.StatusNotFound {
		// TODO: in case of 404 nothing is return by the API
		resp.State.RemoveResource(ctx)
		return
	}

	if itemResponse.StatusCode != http.StatusOK {

		var message tag_struct.ErrorMessage
		if err := json.Unmarshal(resBody, &message); err != nil { // Parse []byte to go struct pointer
			resp.Diagnostics.AddError(
				"Can not unmarshal JSON",
				err.Error(),
			)
			return
		}
		resp.Diagnostics.AddError(
			message.Title,
			message.Detail,
		)
		resp.State.RemoveResource(ctx)
		return
	}

	var result sifflet.TagDto
	if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
		resp.Diagnostics.AddError(
			"Can not unmarshal JSON",
			err.Error(),
		)
		return
	}

	Type := tag_struct.TagDtoType(result.Type)
	state = tag_struct.TagDto{
		CreatedBy:        types.StringValue(*result.CreatedBy),
		CreatedDate:      types.StringValue(strconv.FormatInt(*result.CreatedDate, 10)),
		Description:      result.Description,
		Editable:         types.BoolValue(*result.Editable),
		Id:               types.StringValue(result.Id.String()),
		LastModifiedDate: types.StringValue(strconv.FormatInt(*result.LastModifiedDate, 10)),
		ModifiedBy:       types.StringValue(*result.ModifiedBy),
		Name:             &result.Name,
		Type:             &Type,
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// NOT IMPLEMENTED IN OPENAPI CONTRACT
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state tag_struct.TagDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.String()

	tagResponse, _ := r.client.DeleteTag(ctx, uuid.MustParse(id))
	resBody, _ := io.ReadAll(tagResponse.Body)
	tflog.Debug(ctx, "Response "+string(resBody))

	if tagResponse.StatusCode != http.StatusNoContent {
		var message tag_struct.ErrorMessage
		if err := json.Unmarshal(resBody, &message); err != nil { // Parse []byte to go struct pointer
			resp.Diagnostics.AddError(
				"Can not unmarshal JSON",
				err.Error(),
			)
			return
		}
		resp.Diagnostics.AddError(
			message.Title,
			message.Detail,
		)
		resp.State.RemoveResource(ctx)
		return
	}

}

func (r *credentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*httpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *httpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = clients.AlphaClient
}