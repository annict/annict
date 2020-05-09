# frozen_string_literal: true
# == Schema Information
#
# Table name: activities
#
#  id                         :bigint           not null, primary key
#  action                     :string(510)      not null
#  recipient_type             :string(510)      not null
#  resources_count            :integer          default(0), not null
#  trackable_type             :string(510)      not null
#  created_at                 :datetime
#  updated_at                 :datetime
#  episode_id                 :bigint
#  episode_record_id          :bigint
#  multiple_episode_record_id :bigint
#  recipient_id               :bigint           not null
#  status_id                  :bigint
#  trackable_id               :bigint           not null
#  user_id                    :bigint           not null
#  work_id                    :bigint
#  work_record_id             :bigint
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

  has_many :episode_records, dependent: :nullify
  has_many :statuses, dependent: :nullify
  has_many :work_records, dependent: :nullify

  scope :records_and_reviews, -> { with_action(:create_episode_record, :create_work_record, :create_multiple_episode_records) }

  def resources
    case action
    when "create_status"
      statuses
    when "create_episode_record"
      episode_records
    when "create_work_record"
      work_records
    else
      []
    end
  end

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
