# frozen_string_literal: true

# == Schema Information
#
# Table name: likes
#
#  id             :integer          not null, primary key
#  recipient_type :string(510)      not null
#  created_at     :datetime
#  updated_at     :datetime
#  recipient_id   :integer          not null
#  user_id        :integer          not null
#
# Indexes
#
#  likes_user_id_idx  (user_id)
#
# Foreign Keys
#
#  likes_user_id_fk  (user_id => users.id) ON DELETE => cascade
#

class Like < ApplicationRecord
  belongs_to :recipient, polymorphic: true, counter_cache: true
  belongs_to :user
  has_many :notifications,
    foreign_key: :trackable_id,
    foreign_type: :trackable,
    dependent: :destroy

  after_create :save_notification

  private

  def save_notification
    Notification.create do |n|
      n.user        = recipient.user
      n.action_user = user
      n.trackable   = self
      n.action      = "likes.create"
    end
  end
end
