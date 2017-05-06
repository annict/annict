# == Schema Information
#
# Table name: latest_statuses
#
#  id                  :integer          not null, primary key
#  user_id             :integer          not null
#  work_id             :integer          not null
#  next_episode_id     :integer
#  kind                :integer          not null
#  watched_episode_ids :integer          default([]), not null, is an Array
#  position            :integer          default(0), not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#
# Indexes
#
#  index_latest_statuses_on_next_episode_id       (next_episode_id)
#  index_latest_statuses_on_user_id               (user_id)
#  index_latest_statuses_on_user_id_and_position  (user_id,position)
#  index_latest_statuses_on_user_id_and_work_id   (user_id,work_id) UNIQUE
#  index_latest_statuses_on_work_id               (work_id)
#

class LatestStatus < ApplicationRecord
  include StatusCommon

  acts_as_list scope: :user

  belongs_to :next_episode, class_name: "Episode", optional: true

  scope :desiring_to_watch, -> { with_kind(:wanna_watch, :watching, :on_hold) }
  scope :on_hold, -> { with_kind(:on_hold) }
  scope :wanna_watch_and_watching, -> { with_kind(:wanna_watch, :watching) }
  scope :wanna_watch, -> { with_kind(:wanna_watch) }
  scope :watching, -> { with_kind(:watching) }
  scope :has_next_episode, -> { where.not(next_episode_id: nil) }

  after_save :expire_cache
  after_destroy :expire_cache

  def self.count_on(status_kind)
    work_published.with_kind(status_kind).count
  end

  def self.refresh_next_episode(user)
    latest_statuses = user.latest_statuses.includes(:work).with_kind(:watching)
    latest_statuses.find_each do |s|
      s.update_column(:next_episode_id, s.fetch_next_episode&.id)
    end
  end

  def fetch_next_episode
    episode_ids = work.episodes.published.pluck(:id)
    next_episode_ids = episode_ids - watched_episode_ids
    Episode.where(id: next_episode_ids).order(:sort_number).first
  end

  def append_episode(episode)
    episode_ids = watched_episode_ids << episode.id
    self.watched_episode_ids = episode_ids.uniq
    self.next_episode = self.fetch_next_episode
    move_to_top
    self
  end

  def expire_cache
    user.touch(:status_cache_expired_at)
  end
end
