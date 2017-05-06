# frozen_string_literal: true
# == Schema Information
#
# Table name: episodes
#
#  id                    :integer          not null, primary key
#  work_id               :integer          not null
#  number                :string(510)
#  sort_number           :integer          default(0), not null
#  sc_count              :integer
#  title                 :string(510)
#  checkins_count        :integer          default(0), not null
#  created_at            :datetime
#  updated_at            :datetime
#  prev_episode_id       :integer
#  aasm_state            :string           default("published"), not null
#  fetch_syobocal        :boolean          default(FALSE), not null
#  raw_number            :string
#  avg_rating            :float
#  title_ro              :string           default(""), not null
#  title_en              :string           default(""), not null
#  record_comments_count :integer          default(0), not null
#
# Indexes
#
#  episodes_work_id_idx               (work_id)
#  episodes_work_id_sc_count_key      (work_id,sc_count) UNIQUE
#  index_episodes_on_aasm_state       (aasm_state)
#  index_episodes_on_prev_episode_id  (prev_episode_id)
#

class Episode < ApplicationRecord
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(
    number sort_number sc_count title prev_episode_id fetch_syobocal raw_number
    title_ro title_en
  ).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :prev_episode,
    class_name: "Episode",
    foreign_key: :prev_episode_id,
    optional: true
  belongs_to :work, counter_cache: true
  has_many :activities, dependent: :destroy, as: :recipient
  has_many :records, dependent: :destroy, class_name: "Checkin"
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :draft_episodes, dependent: :destroy
  has_many :programs, dependent: :destroy

  validates :sort_number, presence: true, numericality: { only_integer: true }

  scope :recorded, -> { where("checkins_count > 0") }

  before_create :set_sort_number
  after_create :update_prev_episode
  before_destroy :unset_prev_episode_id
  after_save :expire_cache
  after_destroy :expire_cache

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
    number.blank? && title.present? && title == work.title
  end

  def to_hash
    JSON.parse(to_json)
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field) if send(field).present?
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end

  private

  def unset_prev_episode_id
    return if next_episode.blank?
    # エピソードを削除するとき、次のエピソードの `prev_episode_id` に
    # 削除対象のエピソードが設定されていたとき、その情報を削除する
    next_episode.update_column(:prev_episode_id, nil) if self == next_episode.prev_episode
  end

  def update_prev_episode
    prev_episode = work.episodes.where.not(id: id).order(sort_number: :desc).first
    update_column(:prev_episode_id, prev_episode.id) if prev_episode.present?
  end

  def set_sort_number
    self.sort_number = (work.episodes.count + 1) * 10
  end

  def expire_cache
    programs.update_all(updated_at: Time.now)
  end
end
