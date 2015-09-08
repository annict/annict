class DbActivityDecorator < ApplicationDecorator
  def resource_diff_table
    new_resource = trackable.class.new(parameters["new"])
    old_resource = trackable.class.new(parameters["old"])
    data = {
      resource: trackable,
      diffs: diffs,
      draft_values: new_resource.decorate.to_values,
      origin_values: old_resource.decorate.to_values
    }

    h.render("db/activities/resource_diff_table", data)
  end
end
