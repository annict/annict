class DbActivityDecorator < ApplicationDecorator
  def resource_diff_table
    if action == "multiple_episodes.create"
      data = init_multiple_episodes_data
    else
      new_resource = init_resource(parameters["new"])
      old_resource = init_resource(parameters["old"])

      data = {
        resource: trackable,
        diffs: diffs,
        draft_values: new_resource.decorate.to_values,
        origin_values: old_resource.decorate.to_values
      }
    end

    h.render("db/activities/resource_diff_table", data)
  end

  private

  def init_resource(params)
    resource = trackable.class.new(params)

    case trackable_type
    when "Episode", "Program"
      resource.work = trackable.work
    end

    resource
  end

  def init_multiple_episodes_data
    body = parameters["new"].map { |p| "#{p['number']},#{p['title']}" }.join("\n")
    multiple_episodes = DraftMultipleEpisode.new(body: body)

    {
      resource: multiple_episodes,
      diffs: HashDiff.diff({}, multiple_episodes.to_diffable_hash),
      draft_values: multiple_episodes.decorate.to_values,
      origin_values: {}
    }
  end
end
