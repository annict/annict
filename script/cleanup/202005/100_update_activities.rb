# frozen_string_literal: true

users = User.only_kept
# For test
# users = User.find_by(username: "shimbaco").followings

users.find_each do |user|
  activities = user
    .activities
    .with_action(:create_status, :create_episode_record, :create_work_record)
    .where(activity_group_id: nil)
    .order(:created_at)

  activities.each do |activity|
    puts "----- User: #{user.id}, Activity: #{activity.id}"

    ActiveRecord::Base.transaction do
      activity_group = user.create_or_last_activity_group_for_init!(activity.trackable)
      activity.update_column(:activity_group_id, activity_group.id)
    end
  end
end
