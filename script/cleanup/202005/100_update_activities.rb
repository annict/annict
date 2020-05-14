# frozen_string_literal: true

users = User.only_kept
# For test
# users = User.find_by(username: "shimbaco").followings

users.find_each do |user|
  activities = user.
    activities.
    with_action(:create_status, :create_episode_record, :create_work_record).
    order(:created_at)

  activities.each do |activity|
    puts "----- User: #{user.id}, Activity: #{activity.id}"

    next if activity.activity_group_id

    ActiveRecord::Base.transaction do
      activity_group = user.create_or_last_activity_group!(activity.trackable)

      case activity.trackable_type
      when "EpisodeRecord"
        episode_record = activity.trackable

        activity.update_columns(
          activity_group_id: activity_group.id,
          episode_id: episode_record.episode_id,
          work_id: episode_record.work_id,
          episode_record_id: episode_record.id,
          activity_type: :episode_record
        )
      when "WorkRecord"
        work_record = activity.trackable

        activity.update_columns(
          activity_group_id: activity_group.id,
          work_id: work_record.work_id,
          work_record_id: work_record.id,
          activity_type: :work_record
        )
      when "Status"
        status = activity.trackable

        activity.update_columns(
          activity_group_id: activity_group.id,
          work_id: status.work_id,
          status_id: status.id,
          activity_type: :status
        )
      end
    end
  end
end
