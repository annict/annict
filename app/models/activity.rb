# frozen_string_literal: true

# == Schema Information
#
# Table name: activities
#
#  id                :bigint           not null, primary key
#  itemable_type     :string
#  created_at        :datetime
#  updated_at        :datetime
#  activity_group_id :bigint           not null
#  itemable_id       :bigint
#  user_id           :bigint           not null
#
# Indexes
#
#  activities_user_id_idx                                (user_id)
#  index_activities_on_activity_group_id                 (activity_group_id)
#  index_activities_on_activity_group_id_and_created_at  (activity_group_id,created_at)
#  index_activities_on_created_at                        (created_at)
#  index_activities_on_episode_id                        (episode_id)
#  index_activities_on_episode_record_id                 (episode_record_id)
#  index_activities_on_itemable_id_and_itemable_type     (itemable_id,itemable_type)
#  index_activities_on_multiple_episode_record_id        (multiple_episode_record_id)
#  index_activities_on_status_id                         (status_id)
#  index_activities_on_trackable_id_and_trackable_type   (trackable_id,trackable_type)
#  index_activities_on_work_id                           (work_id)
#  index_activities_on_work_record_id                    (work_record_id)
#
# Foreign Keys
#
#  activities_user_id_fk  (user_id => users.id) ON DELETE => cascade
#  fk_rails_...           (activity_group_id => activity_groups.id)
#  fk_rails_...           (episode_id => episodes.id)
#  fk_rails_...           (episode_record_id => episode_records.id)
#  fk_rails_...           (multiple_episode_record_id => multiple_episode_records.id)
#  fk_rails_...           (status_id => statuses.id)
#  fk_rails_...           (work_id => works.id)
#  fk_rails_...           (work_record_id => work_records.id)
#

class Activity < ApplicationRecord
  extend Enumerize

  self.ignored_columns = %w[
    action
    episode_id
    episode_record_id
    mer_processed_at
    migrated_at
    multiple_episode_record_id
    recipient_id
    recipient_type
    status_id
    trackable_id
    trackable_type
    work_id
    work_record_id
  ]

  enumerize :trackable_type, in: ActivityGroup::ITEMABLE_TYPES, scope: true

  counter_culture :activity_group

  belongs_to :activity_group
  belongs_to :itemable, polymorphic: true
  belongs_to :user

  after_destroy :destroy_activity_group

  # @deprecated
  def action
    "create_#{itemable_type.underscore}"
  end

  # @deprecated
  def deprecated_action
    case action
    when "create_episode_record"
      "create_record"
    when "create_work_record", "create_anime_record"
      "create_review"
    when "create_multiple_episode_records"
      "create_multiple_records"
    else
      action
    end
  end

  private

  def destroy_activity_group
    unless activity_group.activities.exists?
      activity_group.destroy
    end
  end
end
