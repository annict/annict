# == Schema Information
#
# Table name: follows
#
#  id           :integer          not null, primary key
#  user_id      :integer          not null
#  following_id :integer          not null
#  created_at   :datetime
#  updated_at   :datetime
#
# Indexes
#
#  index_follows_on_user_id_and_following_id  (user_id,following_id) UNIQUE
#

class Follow < ActiveRecord::Base
  belongs_to :following, class_name: 'User'
  belongs_to :user

  after_create  :save_notification
  after_destroy :delete_notification

  private

  def save_notification
    Notification.create do |n|
      n.user        = following
      n.action_user = user
      n.trackable   = self
      n.action      = 'follows.create'
    end
  end

  def delete_notification
    Notification.where(trackable_type: 'Follow', trackable_id: id).destroy_all
  end
end
