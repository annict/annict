# frozen_string_literal: true

class DbActivityDecorator < ApplicationDecorator
  def resource_diff_table
    return if action_table_name == "comments"

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

    h.render("db/activities/resource_diff_table", data)
  end
end
