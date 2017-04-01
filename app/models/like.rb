# frozen_string_literal: true
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
  has_many :notifications,
    foreign_key: :trackable_id,
    foreign_type: :trackable,
    dependent: :destroy

  after_create :save_notification
  after_save :expire_cache
  after_destroy :expire_cache

  private

  def save_notification
    Notification.create do |n|
      n.user        = recipient.user
      n.action_user = user
      n.trackable   = self
      n.action      = "likes.create"
    end
  end

  def expire_cache
    return unless recipient_type.in?(%w(Checkin MultipleRecord Status))
    recipient.activities.update_all(updated_at: Time.now)
  end
end
