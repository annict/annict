# frozen_string_literal: true

# == Schema Information
#
# Table name: comments
#
#  id                :bigint           not null, primary key
#  body              :text             not null
#  likes_count       :integer          default(0), not null
#  locale            :string           default("other"), not null
#  created_at        :timestamptz
#  updated_at        :timestamptz
#  episode_record_id :bigint           not null
#  user_id           :bigint           not null
#  work_id           :bigint
#
# Indexes
#
#  comments_checkin_id_idx    (episode_record_id)
#  comments_user_id_idx       (user_id)
#  index_comments_on_locale   (locale)
#  index_comments_on_work_id  (work_id)
#
# Foreign Keys
#
#  comments_checkin_id_fk  (episode_record_id => episode_records.id) ON DELETE => cascade
#  comments_user_id_fk     (user_id => users.id) ON DELETE => cascade
#  fk_rails_...            (work_id => works.id)
#

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
