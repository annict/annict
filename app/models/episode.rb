# frozen_string_literal: true
# == Schema Information
#
# Table name: episodes
#
#  id                          :integer          not null, primary key
#  aasm_state                  :string           default("published"), not null
#  deleted_at                  :datetime
#  episode_record_bodies_count :integer          default(0), not null
#  episode_records_count       :integer          default(0), not null
#  fetch_syobocal              :boolean          default(FALSE), not null
#  number                      :string(510)
#  number_en                   :string           default(""), not null
#  ratings_count               :integer          default(0), not null
#  raw_number                  :float
#  satisfaction_rate           :float
#  sc_count                    :integer
#  score                       :float
#  sort_number                 :integer          default(0), not null
#  title                       :string(510)
#  title_en                    :string           default(""), not null
#  title_ro                    :string           default(""), not null
#  created_at                  :datetime
#  updated_at                  :datetime
#  prev_episode_id             :integer
#  work_id                     :integer          not null
#
# Indexes
#
#  episodes_work_id_idx                                   (work_id)
#  episodes_work_id_sc_count_key                          (work_id,sc_count) UNIQUE
#  index_episodes_on_aasm_state                           (aasm_state)
#  index_episodes_on_deleted_at                           (deleted_at)
#  index_episodes_on_prev_episode_id                      (prev_episode_id)
#  index_episodes_on_ratings_count                        (ratings_count)
#  index_episodes_on_satisfaction_rate                    (satisfaction_rate)
#  index_episodes_on_satisfaction_rate_and_ratings_count  (satisfaction_rate,ratings_count)
#  index_episodes_on_score                                (score)
#
# Foreign Keys
#
#  episodes_work_id_fk  (work_id => works.id) ON DELETE => cascade
#  fk_rails_...         (prev_episode_id => episodes.id)
#

class Episode < ApplicationRecord
  include AASM
  include DbActivityMethods
  include SoftDeletable

  DIFF_FIELDS = %i(
    number sort_number sc_count title prev_episode_id fetch_syobocal raw_number title_en
  ).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  counter_culture :work, column_name: proc { |model| model.published? ? "auto_episodes_count" : nil }

  belongs_to :prev_episode,
    class_name: "Episode",
    foreign_key: :prev_episode_id,
    optional: true
  belongs_to :work
  has_many :activities, dependent: :destroy, as: :recipient
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :episode_records, dependent: :destroy
  has_many :slots, dependent: :nullify

  validates :sort_number, presence: true, numericality: { only_integer: true }

  scope :recorded, -> { where("episode_records_count > 0") }

  after_create :update_prev_episode
  before_destroy :unset_prev_episode_id

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

  def records_chart_dataset
    current_month = Date.today.beginning_of_month
    count = 0
    [
      current_month.months_ago(3),
      current_month.months_ago(2),
      current_month.months_ago(1),
      current_month
    ].map do |date|
      count += episode_records.by_month(date).count
      {
        date: date.to_time.to_datetime.strftime("%Y/%m/%d"),
        value: count
      }
    end.to_json
  end

  def rating_state_chart_dataset
    all_records_count = episode_records.where.not(rating_state: nil).count
    EpisodeRecord.rating_state.values.map do |state|
      state_records_count = episode_records.with_rating_state(state).count
      ratio = state_records_count / all_records_count.to_f
      {
        name: state.text,
        name_key: state,
        quantity: state_records_count,
        percentage: ratio.nan? ? 0 : (ratio * 100).round
      }
    end.to_json
  end

  def local_number
    return number if I18n.locale == :ja
    return "##{raw_number}" if raw_number

    number
  end

  private

  def unset_prev_episode_id
    return if next_episode.nil?

    # エピソードを削除するとき、次のエピソードの `prev_episode_id` に
    # 削除対象のエピソードが設定されていたとき、その情報を削除する
    next_episode.update_column(:prev_episode_id, nil) if self == next_episode.prev_episode
  end

  def update_prev_episode
    prev_episode = work.episodes.where.not(id: id).order(sort_number: :desc).first
    update_column(:prev_episode_id, prev_episode.id) if prev_episode
  end
end
