# == Schema Information
#
# Table name: follows
#
#  id           :bigint           not null, primary key
#  created_at   :timestamptz
#  updated_at   :timestamptz
#  following_id :bigint           not null
#  user_id      :bigint           not null
#
# Indexes
#
#  follows_following_id_idx          (following_id)
#  follows_user_id_following_id_key  (user_id,following_id) UNIQUE
#  follows_user_id_idx               (user_id)
#
# Foreign Keys
#
#  follows_following_id_fk  (following_id => users.id) ON DELETE => cascade
#  follows_user_id_fk       (user_id => users.id) ON DELETE => cascade
#

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
