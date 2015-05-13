module EditRequestsHelper
  def resource_diff_table(edit_request)
    draft_resource_params = edit_request.to_diffable_draft_resource_hash
    resource = if edit_request.resource.present?
      edit_request.resource.to_diffable_hash
    else
      {}
    end

    options = {
      diff: HashDiff.diff(resource, draft_resource_params)
    }

    render("db/application/resource_diff_table", options)
  end
end
