# frozen_string_literal: true
# == Schema Information
#
# Table name: activities
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_id   :integer          not null
#  recipient_type :string           not null
#  trackable_id   :integer          not null
#  trackable_type :string           not null
#  action         :string           not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  index_activities_on_recipient_id_and_recipient_type  (recipient_id,recipient_type)
#  index_activities_on_trackable_id_and_trackable_type  (trackable_id,trackable_type)
#

class Activity < ActiveRecord::Base
  extend Enumerize

  enumerize :action, in: %w(
    create_status
    create_record
    create_multiple_records
  )

  belongs_to :recipient, polymorphic: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user
end
