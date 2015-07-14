# == Schema Information
#
# Table name: episodes
#
#  id              :integer          not null, primary key
#  work_id         :integer          not null
#  number          :string(510)
#  sort_number     :integer          default(0), not null
#  sc_count        :integer
#  title           :string(510)
#  checkins_count  :integer          default(0), not null
#  created_at      :datetime
#  updated_at      :datetime
#  next_episode_id :integer
#  prev_episode_id :integer
#
# Indexes
#
#  episodes_work_id_idx               (work_id)
#  episodes_work_id_sc_count_key      (work_id,sc_count) UNIQUE
#  index_episodes_on_next_episode_id  (next_episode_id)
#  index_episodes_on_prev_episode_id  (prev_episode_id)
#

class Episode < ActiveRecord::Base
  include EpisodeCommon

  has_paper_trail only: DIFF_FIELDS

  belongs_to :old_next_episode, class_name: "Episode", foreign_key: :next_episode_id
  belongs_to :prev_episode, class_name: "Episode", foreign_key: :prev_episode_id
  belongs_to :work, counter_cache: true
  has_many :activities, dependent: :destroy, foreign_key: :recipient_id, foreign_type: :recipient
  has_many :checkins,   dependent: :destroy
  has_many :checks,     dependent: :destroy
  has_many :draft_episodes, dependent: :destroy
  has_many :programs,   dependent: :destroy

  after_create :update_prev_episode
  before_destroy :unset_next_id_on_prev_episode

  def self.create_from_multiple_episodes(work, multiple_episodes)
    episodes_count = work.episodes.count
    multiple_episodes.each do |episode|
      episodes_count += 1
      work.episodes.create do |e|
        e.number = episode[:number]
        e.sort_number = episodes_count * 10
        e.title = episode[:title]
      end
    end
  end

  def old_prev_episode
    work.episodes.find_by(old_next_episode: self)
  end

  def next_episode
    work.episodes.find_by(prev_episode: self)
  end

  def number_title
    "#{number}「#{title}」"
  end

  # 映画やOVAなどの実質エピソードを持たない作品かどうかを判定する
  def single?
    number.blank? && title.present?
  end

  private

  def unset_next_id_on_prev_episode
    if prev_episode.present? && (self == prev_episode.next_episode)
      prev_episode.update_column(:next_episode_id, nil)
    end
  end

  def update_prev_episode
    prev_episode = work.episodes.where.not(id: id).order(sort_number: :desc).first
    update_column(:prev_episode_id, prev_episode.id) if prev_episode.present?
  end
end
