# frozen_string_literal: true
# == Schema Information
#
# Table name: activities
#
#  id              :bigint           not null, primary key
#  action          :string(510)      not null
#  recipient_type  :string(510)      not null
#  repetitiveness  :boolean          default(FALSE), not null
#  resources_count :integer          default(0), not null
#  single          :boolean          default(FALSE), not null
#  trackable_type  :string(510)      not null
#  created_at      :datetime
#  updated_at      :datetime
#  recipient_id    :bigint           not null
#  trackable_id    :bigint           not null
#  user_id         :bigint           not null
#
# Indexes
#
#  activities_user_id_idx                                (user_id)
#  index_activities_on_activity_group_id                 (activity_group_id)
#  index_activities_on_activity_group_id_and_created_at  (activity_group_id,created_at)
#  index_activities_on_created_at                        (created_at)
#  index_activities_on_episode_id                        (episode_id)
#  index_activities_on_episode_record_id                 (episode_record_id)
#  index_activities_on_multiple_episode_record_id        (multiple_episode_record_id)
#  index_activities_on_status_id                         (status_id)
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

  self.ignored_columns = %w(work_id episode_id status_id episode_record_id multiple_episode_record_id work_record_id)

  enumerize :trackable_type, in: %w(
    Status
    EpisodeRecord
    WorkRecord
  ), scope: true

  enumerize :action, in: %w(
    create_status
    create_episode_record
    create_work_record
    create_multiple_episode_records
  ), scope: true

  enumerize :activity_type, in: %w(
    status
    episode_record
    work_record
  ), scope: true

  counter_culture :activity_group

  belongs_to :activity_group, optional: true
  belongs_to :episode, optional: true
  belongs_to :multiple_episode_record, optional: true
  belongs_to :recipient, polymorphic: true
  belongs_to :episode_record, optional: true
  belongs_to :work_record, optional: true
  belongs_to :status, optional: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user
  belongs_to :work, optional: true

  def deprecated_action
    case action
    when "create_episode_record"
      "create_record"
    when "create_work_record"
      "create_review"
    when "create_multiple_episode_records"
      "create_multiple_records"
    else
      action
    end
  end
end
