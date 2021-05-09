# frozen_string_literal: true

# == Schema Information
#
# Table name: activity_groups
#
#  id               :bigint           not null, primary key
#  activities_count :integer          default(0), not null
#  itemable_type    :string           not null
#  single           :boolean          default(FALSE), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#  user_id          :bigint           not null
#
# Indexes
#
#  index_activity_groups_on_created_at  (created_at)
#  index_activity_groups_on_user_id     (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#
class ActivityGroup < ApplicationRecord
  extend Enumerize

  include BatchDestroyable

  ITEMABLE_TYPES = %w[
    Status
    EpisodeRecord
    WorkRecord
  ].freeze

  enumerize :itemable_type, in: ITEMABLE_TYPES, scope: true

  belongs_to :user
  has_many :activities, dependent: :destroy
  has_many :ordered_activities, -> { order(created_at: :desc) }, class_name: "Activity"

  define_prelude(:items) do |activity_groups|
    all_activities = Activity.where(activity_group: activity_groups.pluck(:id))
    all_activities_by_activity_group_id = all_activities.group_by(&:activity_group_id)
    all_activities_by_trackable_type = all_activities.group_by(&:trackable_type)
    all_statuses = Status
      .eager_load(anime: :anime_image)
      .where(id: all_activities_by_trackable_type["Status"]&.pluck(:trackable_id))
      .order(created_at: :desc)
    all_episode_records = EpisodeRecord.where(id: all_activities_by_trackable_type["EpisodeRecord"]&.pluck(:trackable_id))
    all_anime_records = AnimeRecord.where(id: all_activities_by_trackable_type["WorkRecord"]&.pluck(:trackable_id))
    all_records = Record
      .preload(:user, episode_record: [:episode, {anime: :anime_image}])
      .where(id: all_episode_records.pluck(:record_id) + all_anime_records.pluck(:record_id))
      .order(created_at: :desc)

    activity_groups.index_with do |activity_group|
      activities = all_activities_by_activity_group_id[activity_group.id]
      trackable_type = activities.first.trackable_type

      case trackable_type
      when "Status"
        all_statuses.find_all { |s| activities.pluck(:trackable_id).include?(s.id) }.first(2)
      when "EpisodeRecord"
        episode_records = all_episode_records.find_all { |er| activities.pluck(:trackable_id).include?(er.id) }
        all_records.find_all { |r| episode_records.pluck(:record_id).include?(r.id) }.first(2)
      when "WorkRecord"
        anime_records = all_anime_records.find_all { |ar| activities.pluck(:trackable_id).include?(ar.id) }
        all_records.find_all { |r| anime_records.pluck(:record_id).include?(r.id) }.first(2)
      end
    end
  end
end
