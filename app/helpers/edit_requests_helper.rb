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

  def episodes_table(edit_request)
    episodes = convert_episodes(edit_request.draft_resource_params["episodes"])
    render("db/edit_episodes_requests/table", { episodes: episodes })
  end

  def convert_episodes(episodes)
    escaped_episodes = episodes.gsub(/([^\\])\"/, %q/\\1__double_quote__/)

    CSV.parse(escaped_episodes).map do |ary|
      title = ary[1].gsub("__double_quote__", '"') if ary[1].present?
      [ary[0], title]
    end
  end
end
