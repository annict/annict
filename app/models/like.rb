# == Schema Information
#
# Table name: likes
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  recipient_id   :integer          not null
#  recipient_type :string           not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  index_likes_on_recipient_id_and_recipient_type  (recipient_id,recipient_type)
#

class Like < ActiveRecord::Base
  belongs_to :recipient, polymorphic: true, counter_cache: true
  belongs_to :user

  after_create  :save_notification
  after_destroy :delete_notification


  private

  def save_notification
    Notification.create do |n|
      n.user        = recipient.user
      n.action_user = user
      n.trackable   = self
      n.action      = 'likes.create'
    end
  end

  def delete_notification
    Notification.where(trackable_type: 'Like', trackable_id: id).destroy_all
  end
end
