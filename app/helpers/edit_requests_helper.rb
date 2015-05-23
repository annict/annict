module EditRequestsHelper
  def resource_diff_table(edit_request)
    resource = if edit_request.resource.present?
      edit_request.resource.decorate.to_diffable_resource
    else
      {}
    end
    draft_resource = edit_request.decorate.to_diffable_draft_resource

    data = {}
    value = {}
    draft_data = {}
    draft_value = {}

    resource.each do |key, val|
      data[key] = val[:data]
      value[key] = val[:value]
    end

    draft_resource.each do |key, val|
      draft_data[key] = val[:data]
      draft_value[key] = val[:value]
    end

    options = {
      diff: HashDiff.diff(data, draft_data),
      value: value,
      draft_value: draft_value
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

  def translated_key(key)
    case key
    when "channel_id" then "チャンネル"
    when "episode_id" then "エピソード"
    when "started_at" then "放送開始日時"
    end
  end
end
