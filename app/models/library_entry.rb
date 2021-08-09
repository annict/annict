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
#  program_id          :bigint
#  status_id           :bigint
#  user_id             :bigint           not null
#  work_id             :bigint           not null
#
# Indexes
#
#  index_library_entries_on_next_episode_id         (next_episode_id)
#  index_library_entries_on_program_id              (program_id)
#  index_library_entries_on_status_id               (status_id)
#  index_library_entries_on_user_id                 (user_id)
#  index_library_entries_on_user_id_and_position    (user_id,position)
#  index_library_entries_on_user_id_and_program_id  (user_id,program_id) UNIQUE
#  index_library_entries_on_user_id_and_work_id     (user_id,work_id) UNIQUE
#  index_library_entries_on_work_id                 (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (next_episode_id => episodes.id)
#  fk_rails_...  (program_id => programs.id)
#  fk_rails_...  (status_id => statuses.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class LibraryEntry < ApplicationRecord
  acts_as_list scope: :user

  self.ignored_columns = %w[kind]

  belongs_to :next_episode, class_name: "Episode", optional: true
  belongs_to :program, optional: true
  belongs_to :status, optional: true
  belongs_to :user
  belongs_to :anime, foreign_key: :work_id

  scope :desiring_to_watch, -> { with_status(:wanna_watch, :watching, :on_hold) }
  scope :finished_to_watch, -> { with_status(:watched, :stop_watching) }
  scope :on_hold, -> { with_status(:on_hold) }
  scope :positive, -> { with_status(:wanna_watch, :watching, :watched) }
  scope :wanna_watch_and_watching, -> { with_status(:wanna_watch, :watching) }
  scope :wanna_watch, -> { with_status(:wanna_watch) }
  scope :watching, -> { with_status(:watching) }
  scope :has_next_episode, -> { where.not(next_episode_id: nil) }
  scope :with_status, ->(*status_kinds) { joins(:status).where(statuses: {kind: status_kinds}) }
  scope :with_not_deleted_anime, -> { joins(:anime).merge(Anime.only_kept) }

  def self.count_on(status_kind)
    with_not_deleted_anime.with_status(status_kind).count
  end

  def self.status_kinds
    joins(:status)
      .select("library_entries.work_id, statuses.kind as status_kind")
      .each_with_object({}) do |le, h|
      h[le.work_id] = Status.kind_v2_to_v3(Status.kind.find_value(le.status_kind))
    end
  end

  def append_episode!(episode)
    ActiveRecord::Base.transaction do
      new_watched_episode_ids = (watched_episode_ids << episode.id).uniq

      episode_ids = anime.episodes.only_kept.pluck(:id)
      unwatched_episode_ids = episode_ids - new_watched_episode_ids
      next_episode = Episode.only_kept.where(id: unwatched_episode_ids).order(:sort_number).first

      self.watched_episode_ids = new_watched_episode_ids
      self.next_episode = next_episode

      save!
      move_to_top
    end

    self
  end

  def remove_episode!(episode)
    self.watched_episode_ids = watched_episode_ids - [episode.id]
    save!
    self
  end
end
