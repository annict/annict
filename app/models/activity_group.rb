# typed: false
# frozen_string_literal: true

class ActivityGroup < ApplicationRecord
  extend Enumerize

  include BatchDestroyable

  ITEMABLE_TYPES = %w[
    AnimeRecord
    Status
    EpisodeRecord
    WorkRecord
  ].freeze

  enumerize :itemable_type, in: ITEMABLE_TYPES, scope: true

  belongs_to :user
  has_many :activities, dependent: :destroy
  has_many :ordered_activities, -> { order(created_at: :desc) }, class_name: "Activity"

  define_prelude(:first_item) do |activity_groups|
    load_items(activity_groups, :first)
  end

  define_prelude(:items) do |activity_groups|
    load_items(activity_groups, :all)
  end

  def self.load_items(activity_groups, limit)
    all_activities = Activity.where(activity_group: activity_groups.pluck(:id))
    all_activities_by_activity_group_id = all_activities.group_by(&:activity_group_id)
    all_activities_by_trackable_type = all_activities.group_by(&:trackable_type)
    all_statuses = Status
      .eager_load(work: :work_image)
      .where(id: all_activities_by_trackable_type["Status"]&.pluck(:trackable_id))
      .order(created_at: :desc)
    all_episode_records = EpisodeRecord.where(id: all_activities_by_trackable_type["EpisodeRecord"]&.pluck(:trackable_id))
    work_record_ids = (all_activities_by_trackable_type["WorkRecord"]&.pluck(:trackable_id).presence || []) +
      (all_activities_by_trackable_type["WorkRecord"]&.pluck(:trackable_id).presence || [])
    all_work_records = WorkRecord.where(id: work_record_ids)
    all_records = Record
      .preload(:user, :work_record, work: :work_image, episode_record: [:episode])
      .where(id: all_episode_records.pluck(:record_id) + all_work_records.pluck(:record_id))
      .order(created_at: :desc)

    activity_groups.index_with do |activity_group|
      activities = all_activities_by_activity_group_id[activity_group.id]
      trackable_type = activities.first.trackable_type

      items = case trackable_type
      when "Status"
        all_statuses.find_all { |s| activities.pluck(:trackable_id).include?(s.id) }
      when "EpisodeRecord"
        trackable_ids = activities.pluck(:trackable_id)
        episode_records = all_episode_records.where(id: trackable_ids)
        episode_record_ids = episode_records.pluck(:record_id)
        all_records.where(id: episode_record_ids)
      when "AnimeRecord", "WorkRecord"
        work_records = all_work_records.find_all { |ar| activities.pluck(:trackable_id).include?(ar.id) }
        all_records.find_all { |r| work_records.pluck(:record_id).include?(r.id) }
      end

      case limit
      when :first
        items.first
      else
        items
      end
    end
  end
end
