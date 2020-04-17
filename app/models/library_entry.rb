# frozen_string_literal: true
# == Schema Information
#
# Table name: library_entries
#
#  id                  :bigint           not null, primary key
#  position            :integer          default(0), not null
#  watched_episode_ids :bigint           default([]), not null, is an Array
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  next_episode_id     :bigint
#  status_id           :bigint
#  user_id             :bigint           not null
#  work_id             :bigint           not null
#
# Indexes
#
#  index_library_entries_on_next_episode_id       (next_episode_id)
#  index_library_entries_on_status_id             (status_id)
#  index_library_entries_on_user_id               (user_id)
#  index_library_entries_on_user_id_and_position  (user_id,position)
#  index_library_entries_on_user_id_and_work_id   (user_id,work_id) UNIQUE
#  index_library_entries_on_work_id               (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (next_episode_id => episodes.id)
#  fk_rails_...  (status_id => statuses.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class LibraryEntry < ApplicationRecord
  acts_as_list scope: :user

  self.ignored_columns = %w(kind)

  belongs_to :next_episode, class_name: "Episode", optional: true
  belongs_to :status, optional: true
  belongs_to :user
  belongs_to :work

  scope :desiring_to_watch, -> { with_status(:wanna_watch, :watching, :on_hold) }
  scope :finished_to_watch, -> { with_status(:watched, :stop_watching) }
  scope :on_hold, -> { with_status(:on_hold) }
  scope :positive, -> { with_status(:wanna_watch, :watching, :watched) }
  scope :wanna_watch_and_watching, -> { with_status(:wanna_watch, :watching) }
  scope :wanna_watch, -> { with_status(:wanna_watch) }
  scope :watching, -> { with_status(:watching) }
  scope :has_next_episode, -> { where.not(next_episode_id: nil) }
  scope :with_status, ->(*status_kinds) { joins(:status).where(statuses: { kind: status_kinds }) }
  scope :with_not_deleted_work, -> { joins(:work).merge(Work.only_kept) }

  def self.count_on(status_kind)
    with_not_deleted_work.with_status(status_kind).count
  end

  def self.refresh_next_episode(user)
    library_entries = user.library_entries.includes(:work).with_status(:watching)
    next_episode_data = library_entries.fetch_next_episode_data

    library_entries.find_each do |ls|
      next_episode = next_episode_data.detect { |ne| ne[:work_id] == ls.work_id }
      next if ls.next_episode_id == next_episode[:next_episode_id]
      ls.update_column(:next_episode_id, next_episode[:next_episode_id])
    end
  end

  def self.fetch_next_episode_data
    work_ids = pluck(:work_id)

    return [] if work_ids.empty?

    watched_episode_ids = pluck(:watched_episode_ids).flatten

    episode_condition = watched_episode_ids.empty? ? "" : "id NOT IN (#{watched_episode_ids.join(',')}) AND"
    sql = <<~SQL
      WITH ranked_episodes AS (
        SELECT
          id,
          work_id,
          dense_rank() OVER (
            PARTITION BY work_id ORDER BY sort_number ASC
          ) AS episode_rank
        FROM episodes
        WHERE
          #{episode_condition}
          work_id IN (#{work_ids.join(',')}) AND
          deleted_at IS NULL
      )
      SELECT work_id, id FROM ranked_episodes WHERE episode_rank = 1;
    SQL
    next_episodes = Episode.find_by_sql(sql)

    work_ids.map do |work_id|
      next_episode = next_episodes.select { |e| e.work_id == work_id }.first
      {
        work_id: work_id,
        next_episode_id: next_episode&.id
      }
    end
  end

  def fetch_next_episode
    episode_ids = work.episodes.only_kept.pluck(:id)
    next_episode_ids = episode_ids - watched_episode_ids
    Episode.where(id: next_episode_ids).order(:sort_number).first
  end

  def append_episode!(episode)
    ActiveRecord::Base.transaction do
      episode_ids = watched_episode_ids << episode.id
      self.watched_episode_ids = episode_ids.uniq
      self.next_episode = fetch_next_episode
      save!
      move_to_top
    end

    self
  end
end
