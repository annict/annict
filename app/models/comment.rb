# frozen_string_literal: true
# == Schema Information
#
# Table name: comments
#
#  id                :integer          not null, primary key
#  user_id           :integer          not null
#  episode_record_id :integer          not null
#  body              :text             not null
#  likes_count       :integer          default(0), not null
#  created_at        :datetime
#  updated_at        :datetime
#  work_id           :integer
#  locale            :string           default("other"), not null
#
# Indexes
#
#  comments_checkin_id_idx    (episode_record_id)
#  comments_user_id_idx       (user_id)
#  index_comments_on_locale   (locale)
#  index_comments_on_work_id  (work_id)
#

class Comment < ApplicationRecord
  include Localizable

  belongs_to :episode_record, counter_cache: true
  belongs_to :user
  belongs_to :work
  has_many :likes,
    foreign_key: :recipient_id,
    foreign_type: :recipient,
    dependent: :destroy
  has_many :notifications,
    foreign_key: :trackable_id,
    foreign_type: :trackable,
    dependent: :destroy

  validates :body, presence: true, length: { maximum: 500 }

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
