# frozen_string_literal: true
# == Schema Information
#
# Table name: works
#
#  id                   :integer          not null, primary key
#  season_id            :integer
#  sc_tid               :integer
#  title                :string(510)      not null
#  media                :integer          not null
#  official_site_url    :string(510)      default(""), not null
#  wikipedia_url        :string(510)      default(""), not null
#  episodes_count       :integer          default(0), not null
#  watchers_count       :integer          default(0), not null
#  released_at          :date
#  created_at           :datetime
#  updated_at           :datetime
#  twitter_username     :string(510)
#  twitter_hashtag      :string(510)
#  released_at_about    :string
#  aasm_state           :string           default("published"), not null
#  number_format_id     :integer
#  title_kana           :string           default(""), not null
#  title_ro             :string           default(""), not null
#  title_en             :string           default(""), not null
#  official_site_en_url :string           default(""), not null
#  wikipedia_en_url     :string           default(""), not null
#  synopsis             :text             default(""), not null
#  synopsis_en          :text             default(""), not null
#  synopsis_source      :string           default(""), not null
#  synopsis_en_source   :string           default(""), not null
#  mal_anime_id         :integer
#
# Indexes
#
#  index_works_on_aasm_state        (aasm_state)
#  index_works_on_number_format_id  (number_format_id)
#  works_sc_tid_key                 (sc_tid) UNIQUE
#  works_season_id_idx              (season_id)
#

class Work < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include WorkCommon

  has_paper_trail only: DIFF_FIELDS

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      after do
        episodes.published.each(&:hide!)
      end

      transitions from: :published, to: :hidden
    end
  end

  has_many :activities, foreign_key: :recipient_id, foreign_type: :recipient, dependent: :destroy
  has_many :casts, dependent: :destroy
  has_many :checkins, dependent: :destroy
  has_many :draft_episodes, dependent: :destroy
  has_many :draft_casts, dependent: :destroy
  has_many :draft_items, dependent: :destroy
  has_many :draft_multiple_episodes, dependent: :destroy
  has_many :draft_programs, dependent: :destroy
  has_many :draft_staffs, dependent: :destroy
  has_many :draft_works, dependent: :destroy
  has_many :episodes, dependent: :destroy
  has_many :latest_statuses, dependent: :destroy
  has_many :organizations, through: :staffs, source: :resource, source_type: "Organization"
  has_many :programs, dependent: :destroy
  has_many :statuses, dependent: :destroy
  has_many :staffs, dependent: :destroy
  has_one :item, dependent: :destroy

  validates :sc_tid, numericality: { only_integer: true }, allow_blank: true,
                     uniqueness: true
  validates :title, presence: true, uniqueness: { conditions: -> { published } }

  scope :by_season, -> (season_slug) {
    return self if season_slug.blank?

    year, name = season_slug.split("-")

    season_conds = (name == "all") ? { year: year } : { year: year, name: name }
    joins(:season).where(seasons: season_conds)
  }

  scope :program_registered, -> {
    work_ids = joins(:programs).merge(Program.where(work_id: all.pluck(:id))).pluck(:id).uniq
    where(id: work_ids)
  }

  scope :checkedin_by, -> (user) {
    joins(
      "INNER JOIN (
        SELECT DISTINCT work_id, MAX(id) AS checkin_id FROM checkins WHERE checkins.user_id = #{user.id} GROUP BY work_id
      ) AS c2 ON works.id = c2.work_id")
  }

  # 作品画像が設定されていない作品
  scope :itemless, -> {
    joins("LEFT OUTER JOIN items ON items.work_id = works.id").where("items.id IS NULL")
  }

  # リリース時期順に並べる
  scope :order_by_season, -> (type = :asc) {
    joins('LEFT OUTER JOIN "seasons" ON "seasons"."id" = "works"."season_id"').
      order("seasons.sort_number #{type} NULLS LAST, works.id #{type}")
  }

  # 作品のエピソード数分の空白文字列が入った配列を返す
  # Chart.jsのx軸のラベルを消すにはこれしか方法がなかったんだ…! たぶん…。
  def chart_labels
    episodes.published.pluck(:id).map { "" }
  end

  def chart_values
    episodes.published.order(:sort_number).pluck(:checkins_count)
  end

  def checkins_count
    chart_values.reduce(&:+).presence || 0
  end

  def comments_count
    episode_ids = episodes.pluck(:id)
    checkins = Checkin.where(episode_id: episode_ids).where('comment != ?', '')

    checkins.count
  end

  def sync_with_syobocal?
    sc_tid.present?
  end

  def syobocal_url
    "http://cal.syoboi.jp/tid/#{sc_tid}"
  end

  def channels
    if episodes.present?
      programs = Program.where(episode_id: episodes.pluck(:id))
      Channel.where(id: programs.pluck(:channel_id).uniq) if programs.present?
    end
  end

  def current_season?
    season.present? && season.slug == ENV["ANNICT_CURRENT_SEASON"]
  end

  def next_season?
    season.present? && season.slug == ENV["ANNICT_NEXT_SEASON"]
  end

  # 映画などのエピソードを持たない作品かどうか
  def single?
    episodes.count == 1 &&
      episodes.first.number.blank? &&
      episodes.first.title == title
  end
end
