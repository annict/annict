# frozen_string_literal: true
# == Schema Information
#
# Table name: activities
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_id   :integer          not null
#  recipient_type :string(510)      not null
#  trackable_id   :integer          not null
#  trackable_type :string(510)      not null
#  action         :string(510)      not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  activities_user_id_idx  (user_id)
#

class Activity < ActiveRecord::Base
  belongs_to :recipient, polymorphic: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user

  def create_record?
    action == "checkins.create"
  end

  def create_status?
    action == "statuses.create"
  end
end
