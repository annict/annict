# frozen_string_literal: true
# == Schema Information
#
# Table name: multiple_records
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  work_id     :integer          not null
#  likes_count :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_multiple_records_on_user_id  (user_id)
#  index_multiple_records_on_work_id  (work_id)
#

class MultipleRecord < ApplicationRecord
  belongs_to :user
  belongs_to :work
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :records, class_name: "Checkin", dependent: :destroy

  validates :user_id, presence: true

  after_create :save_activity

  private

  def save_activity
    Activity.create! do |a|
      a.user      = user
      a.recipient = work
      a.trackable = self
      a.action    = "create_multiple_records"
    end
  end
end
