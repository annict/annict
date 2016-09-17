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
#  prev_episode_id :integer
#  aasm_state      :string           default("published"), not null
#  fetch_syobocal  :boolean          default(FALSE), not null
#  raw_number      :string
#  avg_rating      :float
#
# Indexes
#
#  episodes_work_id_idx               (work_id)
#  episodes_work_id_sc_count_key      (work_id,sc_count) UNIQUE
#  index_episodes_on_aasm_state       (aasm_state)
#  index_episodes_on_prev_episode_id  (prev_episode_id)
#

class Episode < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include EpisodeCommon

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :prev_episode, class_name: "Episode", foreign_key: :prev_episode_id
  belongs_to :work, counter_cache: true
  has_many :activities, dependent: :destroy, foreign_key: :recipient_id, foreign_type: :recipient
  has_many :checkins,   dependent: :destroy
  has_many :draft_episodes, dependent: :destroy
  has_many :programs,   dependent: :destroy

  scope :recorded, -> { where("checkins_count > 0") }

  after_create :update_prev_episode
  before_destroy :unset_prev_episode_id

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

  # エピソードを削除するとき、次のエピソードの `prev_episode_id` や
  # DraftEpisodeの `prev_episode_id` に削除対象のエピソードが設定されていたとき、
  # その情報を削除する
  def unset_prev_episode_id
    # エピソードを削除するとき、次のエピソードの `prev_episode_id` に
    # 削除対象のエピソードが設定されていたとき、その情報を削除する
    if next_episode.present? && (self == next_episode.prev_episode)
      next_episode.update_column(:prev_episode_id, nil)
    end

    DraftEpisode.where(prev_episode_id: id).update_all(prev_episode_id: nil)
  end

  def update_prev_episode
    prev_episode = work.episodes.where.not(id: id).order(sort_number: :desc).first
    update_column(:prev_episode_id, prev_episode.id) if prev_episode.present?
  end
end
