# frozen_string_literal: true

ActivityGroup.find_each do |activity_group|
  puts "activity_group: #{activity_group.id}"

  first_activity = activity_group.activities.order(:created_at).first

  activity_group.update_columns(
    created_at: first_activity.created_at,
    updated_at: first_activity.updated_at
  )
end
