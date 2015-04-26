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
  has_paper_trail

  belongs_to :next_episode, class_name: 'Episode', foreign_key: :next_episode_id
  belongs_to :next_episode, class_name: 'Episode', foreign_key: :next_episode_id
  belongs_to :work, counter_cache: true
  has_many :activities, dependent: :destroy, foreign_key: :recipient_id, foreign_type: :recipient
  has_many :checkins,   dependent: :destroy
  has_many :checks,     dependent: :destroy
  has_many :programs,   dependent: :destroy

  validates :sort_number, presence: true, numericality: { only_integer: true }
  validate :presence_number_or_title

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

  def presence_number_or_title
    if number.blank? && title.blank?
      errors.add(:number_and_title, "は両方入力するか、どちらかを入力してください。")
    end
  end
end
