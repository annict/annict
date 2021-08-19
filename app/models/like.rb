# frozen_string_literal: true

# == Schema Information
#
# Table name: likes
#
#  id            :bigint           not null, primary key
#  likeable_type :string
#  created_at    :datetime
#  updated_at    :datetime
#  likeable_id   :bigint
#  user_id       :bigint           not null
#
# Indexes
#
#  index_likes_on_likeable_id_and_likeable_type  (likeable_id,likeable_type)
#  likes_user_id_idx                             (user_id)
#
# Foreign Keys
#
#  likes_user_id_fk  (user_id => users.id) ON DELETE => cascade
#

class Like < ApplicationRecord
  self.ignored_columns = %w[
    recipient_id
    recipient_type
  ]

  counter_culture :likeable

  belongs_to :likeable, polymorphic: true
  belongs_to :user
  has_many :notifications, as: :trackable, dependent: :destroy

  validates :likeable_type, inclusion: { in: %w[Comment Record Status] }

  after_create :save_notification

  def self.find_by_resource!(resource)
    unless resource.likeable?
      raise Annict::Errors::NotLikeableError
    end

    recipient = case resource
    when Record
      resource.episode_record? ? resource.episode_record : resource.work_record
    else
      resource
    end

    find_by(recipient: recipient)
  end

  def send_notification_to(user)
    unless recipient.is_a?(EpisodeRecord)
      return
    end

    EmailNotificationService.send_email(
      "liked_episode_record",
      user,
      user.id,
      recipient.id
    )
  end

  private

  def save_notification
    return if user.id == recipient.user.id

    Notification.create do |n|
      n.user = recipient.user
      n.action_user = user
      n.trackable = self
      n.action = "likes.create"
    end
  end
end
