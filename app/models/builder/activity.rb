# frozen_string_literal: true

module Builder::Activity
  def self.build(activity_groups)
    all_activities = activity_groups.flat_map { |ag| ag.activities }
    all_status_activities = all_activities.find_all { |activity| activity.trackable_type == "Status" }
    all_episode_record_activities = all_activities.find_all { |activity| activity.trackable_type == "EpisodeRecord" }
    all_anime_record_activities = all_activities.find_all { |activity| activity.trackable_type == "WorkRecord" }
    all_statuses = Status.where(id: all_status_activities.pluck(:trackable_id)).preload(work: :work_image)
    all_records = Record.eager_load(:episode_record, :work_record)
      .merge(EpisodeRecord.where(id: all_episode_record_activities.pluck(:trackable_id)).or(
        AnimeRecord.where(id: all_anime_record_activities.pluck(:trackable_id))
      ))

    activity_groups.map do |activity_group|
      trackable_ids = all_activities
        .find_all { |activity| activity.activity_group_id == activity_group.id }
        .pluck(:trackable_id)

      itemable_type, itemables = case activity_group.itemable_type
      when "Status"
        ["status", all_statuses.find_all { |status| trackable_ids.include?(status.id) }]
      when "EpisodeRecord", "WorkRecord"
        ["record", all_records.find_all { |record| trackable_ids.include?(record.id) }]
      end

      Builder::Activity::ActivityGroupStruct.new(
        itemable_type: itemable_type,
        user: activity_group.user,
        created_at: activity_group.created_at,
        itemables: itemables,
        activities_count: activity_group.activities_count
      )
    end
  end
end
