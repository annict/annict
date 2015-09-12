class DbActivityDecorator < ApplicationDecorator
  def resource_diff_table
    if action == "multiple_episodes.create"
      data = init_multiple_episodes_data
    else
      new_resource = trackable.class.new(parameters["new"])
      old_resource = trackable.class.new(parameters["old"])
      origin_values = if parameters["old"]
        old_resource.decorate.to_values
      else
        {}
      end

      data = {
        resource: trackable,
        diffs: diffs(new_resource, old_resource),
        draft_values: new_resource.decorate.to_values,
        origin_values: origin_values
      }
    end

    h.render("db/activities/resource_diff_table", data)
  end

  private

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
