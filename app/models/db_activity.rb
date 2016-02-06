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
#  parameters     :json
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

  def diffs(new_resource, old_resource)
    HashDiff.diff(old_resource.to_diffable_hash, new_resource.to_diffable_hash)
  end

  private

  def send_notification
    method = case action
      when "edit_request_comments.create" then :comment_notification
      when "edit_requests.publish" then :publish_notification
      when "edit_requests.close" then :close_notification
      end

    if method.present?
      recipient.participants.where.not(user: user).each do |p|
        EditRequestMailer.send(method, id, p.user.email).deliver_later
      end
    end
  end
end
