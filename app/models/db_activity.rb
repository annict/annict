# == Schema Information
#
# Table name: db_activities
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_id   :integer
#  recipient_type :string
#  trackable_id   :integer          not null
#  trackable_type :string           not null
#  action         :string           not null
#  created_at     :datetime         not null
#  updated_at     :datetime         not null
#
# Indexes
#
#  index_db_activities_on_recipient_id_and_recipient_type  (recipient_id,recipient_type)
#  index_db_activities_on_trackable_id_and_trackable_type  (trackable_id,trackable_type)
#

class DbActivity < ActiveRecord::Base
  extend Enumerize

  belongs_to :recipient, polymorphic: true
  belongs_to :trackable, polymorphic: true
  belongs_to :user

  after_create :send_notification

  private

  def send_notification
    case action
    when "edit_request_comments.create"
      EditRequestMailer.comment_notification(id).deliver_later
    when "edit_requests.publish"
      EditRequestMailer.publish_notification(id).deliver_later
    when "edit_requests.close"
      EditRequestMailer.close_notification(id).deliver_later
    end
  end
end
