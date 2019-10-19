# frozen_string_literal: true
# == Schema Information
#
# Table name: activities
#
#  id                         :integer          not null, primary key
#  action                     :string(510)      not null
#  recipient_type             :string(510)      not null
#  trackable_type             :string(510)      not null
#  created_at                 :datetime
#  updated_at                 :datetime
#  episode_id                 :integer
#  episode_record_id          :integer
#  multiple_episode_record_id :integer
#  recipient_id               :integer          not null
#  status_id                  :integer
#  trackable_id               :integer          not null
#  user_id                    :integer          not null
#  work_id                    :integer
#  work_record_id             :integer
#
# Indexes
#
#  activities_user_id_idx                          (user_id)
#  index_activities_on_episode_id                  (episode_id)
#  index_activities_on_episode_record_id           (episode_record_id)
#  index_activities_on_multiple_episode_record_id  (multiple_episode_record_id)
#  index_activities_on_status_id                   (status_id)
#  index_activities_on_work_id                     (work_id)
#  index_activities_on_work_record_id              (work_record_id)
#
# Foreign Keys
#
#  activities_user_id_fk  (user_id => users.id) ON DELETE => cascade
#  fk_rails_...           (episode_id => episodes.id)
#  fk_rails_...           (episode_record_id => episode_records.id)
#  fk_rails_...           (multiple_episode_record_id => multiple_episode_records.id)
#  fk_rails_...           (status_id => statuses.id)
#  fk_rails_...           (work_id => works.id)
#  fk_rails_...           (work_record_id => work_records.id)
#

class Activity < ApplicationRecord
  extend Enumerize

  enumerize :action, in: %w(
    create_status
    create_episode_record
    create_work_record
    create_multiple_episode_records
  ), scope: true

  belongs_to :episode, optional: true
  belongs_to :multiple_episode_record, optional: true
  belongs_to :recipient, polymorphic: true
  belongs_to :episode_record, optional: true
  belongs_to :work_record, optional: true
  belongs_to :status, optional: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user
  belongs_to :work, optional: true

  scope :records_and_reviews, -> { with_action(:create_episode_record, :create_work_record, :create_multiple_episode_records) }

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
