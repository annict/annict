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
#
# Indexes
#
#  episodes_work_id_idx               (work_id)
#  episodes_work_id_sc_count_key      (work_id,sc_count) UNIQUE
#  index_episodes_on_next_episode_id  (next_episode_id)
#

class Episode < ActiveRecord::Base
  include EpisodeCommon

  has_paper_trail

  belongs_to :next_episode, class_name: "Episode", foreign_key: :next_episode_id
  belongs_to :work, counter_cache: true
  has_many :activities, dependent: :destroy, foreign_key: :recipient_id, foreign_type: :recipient
  has_many :checkins,   dependent: :destroy
  has_many :checks,     dependent: :destroy
  has_many :draft_episodes, dependent: :destroy
  has_many :programs,   dependent: :destroy

  after_create :create_nicoch_program
  before_destroy :unset_next_id_on_prev_episode

  def prev_episode
    work.episodes.find_by(next_episode: self)
  end

  def number_title
    "#{number}「#{title}」"
  end

  # 映画やOVAなどの実質エピソードを持たない作品かどうかを判定する
  def single?
    number.blank? && title.present?
  end

  private

  def create_nicoch_program
    if work.broadcast_on_nicoch?
      channel = Channel.find_by(name: 'ニコニコチャンネル')
      nicoch_started_day = (7 * work.episodes.count) - 7
      started_at = work.nicoch_started_at + nicoch_started_day.day

      work.programs.create(channel_id: channel.id, episode_id: id, started_at: started_at)
    end
  end

  def unset_next_id_on_prev_episode
    if prev_episode.present? && (self == prev_episode.next_episode)
      prev_episode.update_column(:next_episode_id, nil)
    end
  end
end
