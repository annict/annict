# == Schema Information
#
# Table name: activities
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_id   :integer          not null
#  recipient_type :string(255)      not null
#  trackable_id   :integer          not null
#  trackable_type :string(255)      not null
#  action         :string(255)      not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  index_activities_on_recipient_id_and_recipient_type  (recipient_id,recipient_type)
#  index_activities_on_trackable_id_and_trackable_type  (trackable_id,trackable_type)
#

class Activity < ActiveRecord::Base
  belongs_to :recipient, polymorphic: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user
end
