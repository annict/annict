# typed: false

class Follow < ApplicationRecord
  counter_culture :following, column_name: :followers_count
  counter_culture :user, column_name: :following_count

  belongs_to :following, class_name: "User"
  belongs_to :user

  after_create :save_notification
  after_destroy :delete_notification

  private

  def save_notification
    Notification.create do |n|
      n.user = following
      n.action_user = user
      n.trackable = self
      n.action = "follows.create"
    end
  end

  def delete_notification
    Notification.where(trackable_type: "Follow", trackable_id: id).destroy_all
  end
end
