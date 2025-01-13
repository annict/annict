# typed: false
# frozen_string_literal: true

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
    work_id
    work_record_id
  ]

  enumerize :trackable_type, in: ActivityGroup::ITEMABLE_TYPES, scope: true

  counter_culture :activity_group

  belongs_to :activity_group
  belongs_to :itemable, foreign_key: :trackable_id, foreign_type: :trackable_type, polymorphic: true
  belongs_to :user

  after_destroy :destroy_activity_group

  def itemable_type
    trackable_type
  end

  def itemable_id
    trackable_id
  end

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
