# frozen_string_literal: true

# == Schema Information
#
# Table name: works
#
#  id                    :integer          not null, primary key
#  season_id             :integer
#  sc_tid                :integer
#  title                 :string(510)      not null
#  media                 :integer          not null
#  official_site_url     :string(510)      default(""), not null
#  wikipedia_url         :string(510)      default(""), not null
#  episodes_count        :integer          default(0), not null
#  watchers_count        :integer          default(0), not null
#  released_at           :date
#  created_at            :datetime
#  updated_at            :datetime
#  twitter_username      :string(510)
#  twitter_hashtag       :string(510)
#  released_at_about     :string
#  aasm_state            :string           default("published"), not null
#  number_format_id      :integer
#  title_kana            :string           default(""), not null
#  title_ro              :string           default(""), not null
#  title_en              :string           default(""), not null
#  official_site_url_en  :string           default(""), not null
#  wikipedia_url_en      :string           default(""), not null
#  synopsis              :text             default(""), not null
#  synopsis_en           :text             default(""), not null
#  synopsis_source       :string           default(""), not null
#  synopsis_source_en    :string           default(""), not null
#  mal_anime_id          :integer
#  facebook_og_image_url :string           default(""), not null
#  twitter_image_url     :string           default(""), not null
#  recommended_image_url :string           default(""), not null
#  season_year           :integer
#  season_name           :integer
#
# Indexes
#
#  index_works_on_aasm_state                   (aasm_state)
#  index_works_on_number_format_id             (number_format_id)
#  index_works_on_season_year                  (season_year)
#  index_works_on_season_year_and_season_name  (season_year,season_name)
#  works_sc_tid_key                            (sc_tid) UNIQUE
#  works_season_id_idx                         (season_id)
#

class Work < ApplicationRecord
  extend Enumerize
  include AASM
  include DbActivityMethods
  include RootResourceCommon

  DIFF_FIELDS = %i(
    sc_tid title title_kana title_en title_ro media official_site_url
    official_site_url_en wikipedia_url wikipedia_url_en twitter_username
    twitter_hashtag number_format_id synopsis synopsis_en synopsis_source
    synopsis_source_en mal_anime_id season_year season_name
  ).freeze

  enumerize :media, in: { tv: 1, ova: 2, movie: 3, web: 4, other: 0 }
  enumerize :season_name, in: Season::NAME_HASH

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

  belongs_to :number_format, optional: true
  belongs_to :season_model,
    class_name: "SeasonModel",
    foreign_key: :season_id,
    optional: true
  has_many :activities,
    foreign_key: :recipient_id,
    foreign_type: :recipient,
    dependent: :destroy
  has_many :casts, dependent: :destroy
  has_many :cast_people, through: :casts, source: :person
  has_many :channel_works, dependent: :destroy
  has_many :characters, through: :casts
  has_many :checkins, dependent: :destroy
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :episodes, dependent: :destroy
  has_many :latest_statuses, dependent: :destroy
  has_many :staffs, dependent: :destroy
  has_many :organizations,
    through: :staffs,
    source: :resource,
    source_type: "Organization"
  has_many :programs, dependent: :destroy
  has_many :series_works, dependent: :destroy
  has_many :series_list, through: :series_works, source: :series
  has_many :statuses, dependent: :destroy
  has_many :staff_people, through: :staffs, source: :resource, source_type: "Person"
  has_one :work_image, dependent: :destroy
  has_one :item, dependent: :destroy

  validates :sc_tid,
    numericality: { only_integer: true },
    allow_blank: true,
    uniqueness: true
  validates :title, presence: true, uniqueness: { conditions: -> { published } }
  validates :media, presence: true
  validates :official_site_url, url: { allow_blank: true }
  validates :official_site_url_en, url: { allow_blank: true }
  validates :wikipedia_url, url: { allow_blank: true }
  validates :wikipedia_url_en, url: { allow_blank: true }
  validates :synopsis, presence_pair: :synopsis_source
  validates :synopsis_en, presence_pair: :synopsis_source_en

  scope(:by_season, ->(season_slug) {
    return self if season_slug.blank?

    where(Season.find_by_slug(season_slug).work_conditions)
  })

  scope(:by_seasons, ->(season_slugs) {
    return self if season_slugs.blank?

    season_pairs = season_slugs.map do |slug|
      season = Season.find_by_slug(slug)
      [season.year, season.name]
    end
    season_year, season_name = season_pairs.shift

    t = Work.arel_table
    works = where(t[:season_year].eq(season_year)).
      where(t[:season_name].eq(season_name))
    season_pairs.inject(works) do |query, season_pair|
      query.
        or(
          where(t[:season_year].eq(season_pair[0])).
          where(t[:season_name].eq(season_pair[1]))
        )
    end
  })

  scope :program_registered, -> {
    work_ids = joins(:programs).
      merge(Program.published.where(work_id: all.pluck(:id))).
      pluck(:id).
      uniq
    where(id: work_ids)
  }

  scope :checkedin_by, -> (user) {
    joins(
      "INNER JOIN (
        SELECT DISTINCT work_id, MAX(id) AS checkin_id FROM checkins
          WHERE checkins.user_id = #{user.id} GROUP BY work_id
      ) AS c2 ON works.id = c2.work_id"
    )
  }

  # 作品画像が設定されていない作品
  scope :image_not_attached, -> {
    joins("LEFT OUTER JOIN work_images ON work_images.work_id = works.id").
      where("work_images.id IS NULL")
  }

  # リリース時期順に並べる
  scope :order_by_season, -> (type = :asc) {
    joins('LEFT OUTER JOIN "seasons" ON "seasons"."id" = "works"."season_id"').
      order("seasons.sort_number #{type} NULLS LAST, works.id #{type}")
  }

  def people
    Person.where(id: (cast_people.pluck(:id) | staff_people.pluck(:id)))
  end

  def season
    return if season_year.blank?
    @season ||= Season.new(season_year, season_name.presence || "all")
  end

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
    checkins = Checkin.where(episode_id: episode_ids).where("comment != ?", "")

    checkins.count
  end

  def sync_with_syobocal?
    sc_tid.present?
  end

  def syobocal_url
    "http://cal.syoboi.jp/tid/#{sc_tid}"
  end

  def twitter_avatar_url(size = :original)
    return "" if twitter_username.blank?
    "https://twitter.com/#{twitter_username}/profile_image?size=#{size}"
  end

  def channels
    return nil if episodes.blank?

    programs = Program.published.where(episode_id: episodes.pluck(:id))
    Channel.published.where(id: programs.pluck(:channel_id).uniq) if programs.present?
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

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :media
        send(field).to_s
      else
        send(field)
      end

      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
