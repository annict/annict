# typed: false
# frozen_string_literal: true

class Comment < ApplicationRecord
  include UgcLocalizable

  counter_culture :episode_record

  belongs_to :episode_record
  belongs_to :user
  belongs_to :work
  has_many :likes, as: :recipient, dependent: :destroy
  has_many :notifications, as: :trackable, dependent: :destroy

  validates :body, presence: true, length: {maximum: 500}

  after_create :save_notification

  private

  def save_notification
    return if episode_record.user == user

    Notification.create do |n|
      n.user = episode_record.user
      n.action_user = user
      n.trackable = self
      n.action = "comments.create"
    end
  end
end
