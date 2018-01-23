# frozen_string_literal: true
# == Schema Information
#
# Table name: activities
#
#  id                 :integer          not null, primary key
#  user_id            :integer          not null
#  recipient_id       :integer          not null
#  recipient_type     :string(510)      not null
#  trackable_id       :integer          not null
#  trackable_type     :string(510)      not null
#  action             :string(510)      not null
#  created_at         :datetime
#  updated_at         :datetime
#  work_id            :integer
#  episode_id         :integer
#  status_id          :integer
#  record_id          :integer
#  multiple_record_id :integer
#  review_id          :integer
#
# Indexes
#
#  activities_user_id_idx                  (user_id)
#  index_activities_on_episode_id          (episode_id)
#  index_activities_on_multiple_record_id  (multiple_record_id)
#  index_activities_on_record_id           (record_id)
#  index_activities_on_review_id           (review_id)
#  index_activities_on_status_id           (status_id)
#  index_activities_on_work_id             (work_id)
#

class Activity < ApplicationRecord
  extend Enumerize

  enumerize :action, in: %w(
    create_status
    create_record
    create_review
    create_multiple_records
  ), scope: true

  belongs_to :episode, optional: true
  belongs_to :multiple_record, optional: true
  belongs_to :recipient, polymorphic: true
  belongs_to :record, class_name: "Checkin", optional: true
  belongs_to :review, optional: true
  belongs_to :status, optional: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user
  belongs_to :work, optional: true

  scope :records_and_reviews, -> { with_action(:create_record, :create_review, :create_multiple_records) }
end
