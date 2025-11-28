# typed: false
# frozen_string_literal: true

class MultipleEpisodeRecord < ApplicationRecord
  belongs_to :user
  belongs_to :work
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :episode_records, dependent: :destroy
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :user_id, presence: true

  after_create :save_activity

  private

  def save_activity
    Activity.create! do |a|
      a.user = user
      a.recipient = work
      a.trackable = self
      a.action = "create_multiple_episode_records"
      a.work = work
      a.multiple_episode_record = self
    end
  end
end
