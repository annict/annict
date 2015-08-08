# == Schema Information
#
# Table name: comments
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  checkin_id  :integer          not null
#  body        :text             not null
#  created_at  :datetime
#  updated_at  :datetime
#  likes_count :integer          default(0), not null
#

class Comment < ActiveRecord::Base
  belongs_to :checkin, counter_cache: true
  belongs_to :user
  has_many   :likes, foreign_key: :recipient_id, foreign_type: :recipient, dependent: :destroy
  has_many   :notifications, foreign_key: :trackable_id, foreign_type: :trackable, dependent: :destroy

  validates :body, presence: true, length: { maximum: 500 }

  after_create  :save_notification


  private

  def save_notification
    if checkin.user != user
      Notification.create do |n|
        n.user        = checkin.user
        n.action_user = user
        n.trackable   = self
        n.action      = 'comments.create'
      end
    end
  end
end
