# frozen_string_literal: true

# == Schema Information
#
# Table name: works
#
#  id                           :bigint           not null, primary key
#  aasm_state                   :string           default("published"), not null
#  deleted_at                   :datetime
#  ended_on                     :date
#  episodes_count               :integer          default(0), not null
#  facebook_og_image_url        :string           default(""), not null
#  manual_episodes_count        :integer
#  media                        :integer          not null
#  no_episodes                  :boolean          default(FALSE), not null
#  official_site_url            :string(510)      default(""), not null
#  official_site_url_en         :string           default(""), not null
#  ratings_count                :integer          default(0), not null
#  recommended_image_url        :string           default(""), not null
#  records_count                :integer          default(0), not null
#  released_at                  :date
#  released_at_about            :string
#  satisfaction_rate            :float
#  sc_tid                       :integer
#  score                        :float
#  season_name                  :integer
#  season_year                  :integer
#  start_episode_raw_number     :float            default(1.0), not null
#  started_on                   :date
#  synopsis                     :text             default(""), not null
#  synopsis_en                  :text             default(""), not null
#  synopsis_source              :string           default(""), not null
#  synopsis_source_en           :string           default(""), not null
#  title                        :string(510)      not null
#  title_alter                  :string           default(""), not null
#  title_alter_en               :string           default(""), not null
#  title_en                     :string           default(""), not null
#  title_kana                   :string           default(""), not null
#  title_ro                     :string           default(""), not null
#  twitter_hashtag              :string(510)
#  twitter_image_url            :string           default(""), not null
#  twitter_username             :string(510)
#  unpublished_at               :datetime
#  watchers_count               :integer          default(0), not null
#  wikipedia_url                :string(510)      default(""), not null
#  wikipedia_url_en             :string           default(""), not null
#  work_records_count           :integer          default(0), not null
#  work_records_with_body_count :integer          default(0), not null
#  created_at                   :datetime
#  updated_at                   :datetime
#  key_pv_id                    :bigint
#  mal_anime_id                 :integer
#  number_format_id             :bigint
#  season_id                    :bigint
#
# Indexes
#
#  index_works_on_aasm_state                           (aasm_state)
#  index_works_on_deleted_at                           (deleted_at)
#  index_works_on_key_pv_id                            (key_pv_id)
#  index_works_on_number_format_id                     (number_format_id)
#  index_works_on_ratings_count                        (ratings_count)
#  index_works_on_satisfaction_rate                    (satisfaction_rate)
#  index_works_on_satisfaction_rate_and_ratings_count  (satisfaction_rate,ratings_count)
#  index_works_on_score                                (score)
#  index_works_on_season_year                          (season_year)
#  index_works_on_season_year_and_season_name          (season_year,season_name)
#  index_works_on_unpublished_at                       (unpublished_at)
#  works_season_id_idx                                 (season_id)
#
# Foreign Keys
#
#  fk_rails_...        (key_pv_id => trailers.id)
#  fk_rails_...        (number_format_id => number_formats.id)
#  works_season_id_fk  (season_id => seasons.id) ON DELETE => cascade
#

class Work < ApplicationRecord
  extend Enumerize

  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[
    sc_tid title title_kana title_en media official_site_url
    official_site_url_en wikipedia_url wikipedia_url_en twitter_username
    twitter_hashtag number_format_id synopsis synopsis_en synopsis_source
    synopsis_source_en mal_anime_id season_year season_name manual_episodes_count
    started_on ended_on
  ].freeze

  attr_accessor :status_kind

  delegate :copyright, to: :work_image, allow_nil: true

  enumerize :media, in: {tv: 1, ova: 2, movie: 3, web: 4, other: 0}
  enumerize :season_name, in: Season::NAME_HASH

  belongs_to :number_format, optional: true
  belongs_to :season_model, class_name: "SeasonModel", foreign_key: :season_id, optional: true
  has_many :casts, dependent: :destroy, foreign_key: :work_id
  has_many :programs, dependent: :destroy, foreign_key: :work_id
  has_many :series_works, dependent: :destroy, foreign_key: :work_id
  has_many :staffs, dependent: :destroy, foreign_key: :work_id
  has_many :work_taggings
  has_many :activities, as: :recipient
  has_many :cast_people, through: :casts, source: :person
  has_many :characters, through: :casts
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :episode_records
  has_many :episodes, dependent: :destroy, foreign_key: :work_id
  has_many :library_entries, foreign_key: :work_id
  has_many :organizations, through: :staffs, source: :resource, source_type: "Organization"
  has_many :slots, dependent: :destroy, foreign_key: :work_id
  has_many :trailers, dependent: :destroy, foreign_key: :work_id
  has_many :records, foreign_key: :work_id
  has_many :series_list, through: :series_works, source: :series
  has_many :statuses
  has_many :staff_people, through: :staffs, source: :resource, source_type: "Person"
  has_many :channels, through: :programs
  has_many :work_records, foreign_key: :work_id
  has_many :records_only_work, class_name: "Record", source: :record, through: :work_records
  has_many :work_tags, through: :work_taggings
  has_one :work_image, dependent: :destroy, foreign_key: :work_id

  validates :media, presence: true
  validates :official_site_url_en, url: {allow_blank: true}
  validates :official_site_url, url: {allow_blank: true}
  validates :sc_tid, numericality: {only_integer: true}, allow_blank: true
  validates :synopsis_en, presence_pair: :synopsis_source_en
  validates :synopsis, presence_pair: :synopsis_source
  validates :title, presence: true, uniqueness: {conditions: -> { only_kept }}
  validates :wikipedia_url_en, url: {allow_blank: true}
  validates :wikipedia_url, url: {allow_blank: true}

  scope(:by_season, ->(season_slug) {
    return self if season_slug.blank?

    where(Season.find_by_slug(season_slug).work_conditions)
  })

  scope(:by_seasons, ->(season_slugs) {
    return self if season_slugs.blank?

    season_pairs = season_slugs.map { |slug|
      season = Season.find_by_slug(slug)
      [season.year, season.name]
    }
    season_year, season_name = season_pairs.shift

    t = Work.arel_table
    works = where(t[:season_year].eq(season_year))
      .where(t[:season_name].eq(season_name))
    season_pairs.inject(works) do |query, season_pair|
      query
        .or(
          where(t[:season_year].eq(season_pair[0]))
          .where(t[:season_name].eq(season_pair[1]))
        )
    end
  })

  scope :slot_registered, -> {
    work_ids = joins(:slots)
      .merge(Slot.only_kept.where(work_id: all.pluck(:id)))
      .pluck(:id)
      .uniq
    where(id: work_ids)
  }

  scope :tracked_by, ->(user) {
    joins(
      "INNER JOIN (
        SELECT DISTINCT work_id, MAX(id) AS record_id FROM records
          WHERE records.user_id = #{user.id} GROUP BY work_id
      ) AS c2 ON works.id = c2.work_id"
    )
  }

  scope :with_no_season, -> {
    where(season_year: nil, season_name: nil)
  }

  scope :with_no_episodes, -> {
    where(no_episodes: false).where(<<~SQL)
      NOT EXISTS (
        SELECT * FROM episodes WHERE
          1 = 1
          AND episodes.work_id = works.id
          AND episodes.deleted_at IS NULL
          AND episodes.unpublished_at IS NULL
      )
    SQL
  }

  scope :with_no_slots, -> {
    where(<<~SQL)
      NOT EXISTS (
        SELECT * FROM slots WHERE
          1 = 1
          AND slots.work_id = works.id
          AND slots.deleted_at IS NULL
          AND slots.unpublished_at IS NULL
      )
    SQL
  }

  # 作品画像が設定されていない作品
  scope :with_no_image, -> {
    joins("LEFT OUTER JOIN work_images ON work_images.work_id = works.id")
      .where("work_images.id IS NULL")
  }

  scope :order_by_season, ->(type = :asc) {
    order(season_year: type, season_name: type)
  }

  scope :season_from, -> (season) {
    where("season_year >= ? AND season_name >= ?", season.year, season.name_value)
      .or(where("season_year > ?", season.year))
      .or(where(season_year: season.year, season_name: nil))
      .or(where(season_year: nil))
  }

  scope :season_until, -> (season) {
    where("season_year <= ? AND season_name <= ?", season.year, season.name_value)
      .or(where("season_year < ?", season.year))
  }

  scope :from_current_season, -> {
    season_from Season.find_by_slug(ENV.fetch("ANNICT_CURRENT_SEASON"))
  }

  scope :until_current_season, -> {
    season_until Season.find_by_slug(ENV.fetch("ANNICT_CURRENT_SEASON"))
  }

  def self.statuses(work_ids, user)
    work_ids = work_ids.uniq
    library_entries = LibraryEntry.where(user: user, work_id: work_ids).eager_load(:status)

    work_ids.map do |work_id|
      {
        work_id: work_id,
        kind: library_entries.find { |ls| ls.work_id == work_id }&.status&.kind.presence || "no_select"
      }
    end
  end

  def people
    Person.where(id: (cast_people.pluck(:id) | staff_people.pluck(:id)))
  end

  def season
    return if season_year.blank?

    @season ||= Season.new(season_year, season_name.presence || "all")
  end

  def build_work_record(
    user:,
    watched_at:,
    rating_overall: nil,
    rating_animation: nil,
    rating_music: nil,
    rating_story: nil,
    rating_character: nil,
    comment: "",
    share_to_twitter: false
  )
    work_record = work_records.new(
      user: user,
      rating_overall_state: rating_overall&.downcase,
      rating_animation_state: rating_animation&.downcase,
      rating_music_state: rating_music&.downcase,
      rating_story_state: rating_story&.downcase,
      rating_character_state: rating_character&.downcase,
      body: comment,
      share_to_twitter: share_to_twitter
    )
    work_record.detect_locale!(:body)
    work_record.build_record(user: user, work: self, watched_at: watched_at)
    work_record
  end

  def comments_count
    episode_ids = episodes.pluck(:id)
    records = Record.where(episode_id: episode_ids).where("comment != ?", "")

    records.count
  end

  def syobocal_tid
    sc_tid
  end

  def sync_with_syobocal?
    sc_tid.present?
  end

  def episodes_filled?
    !manual_episodes_count.nil? && episodes.only_kept.count >= manual_episodes_count
  end

  def slots_exists?
    slots.where.not(started_at: nil).exists?
  end

  def manual_episodes_creatable?
    !episodes_filled? && !slots_exists?
  end

  def syobocal_url
    "http://cal.syoboi.jp/tid/#{sc_tid}"
  end

  def mal_anime_url
    "https://myanimelist.net/anime/#{mal_anime_id}"
  end

  def twitter_avatar_url(size = :original)
    return "" if twitter_username.blank?

    "https://twitter.com/#{twitter_username}/profile_image?size=#{size}"
  end

  def twitter_username_url
    url = "https://twitter.com/#{twitter_username}"
    Addressable::URI.parse(url).normalize.to_s
  end

  def twitter_hashtag_url
    url = "https://twitter.com/search?q=%23#{twitter_hashtag}"
    Addressable::URI.parse(url).normalize.to_s
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

  def duration
    30
  end

  def actual_episodes_count
    manual_episodes_count.presence || episodes_count
  end

  def hashtag_with_hash
    return "" if twitter_hashtag.blank?

    "##{twitter_hashtag}"
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = case field
      when :media
        send(field).to_s
      else
        send(field)
      end

      hash
    }

    data.delete_if { |_, v| v.blank? }
  end

  def watchers_chart_dataset
    (3.months.ago.to_date..Date.today).map { |date|
      count = library_entries.with_status(:wanna_watch, :watching, :watched).before(date).count
      {
        date: date.to_time.to_datetime.strftime("%Y/%m/%d"),
        value: count
      }
    }.to_json
  end

  def status_chart_dataset
    Status.kind.values.map { |kind|
      kind_count = library_entries.with_status(kind).count
      {
        name: kind.text,
        value: kind_count
      }
    }.to_json
  end

  def related_works
    series_work_ids = SeriesWork.where(series_id: series_list.pluck(:id)).pluck(:id)
    series_works = SeriesWork.where(id: series_work_ids)
    Work.where(id: series_works.pluck(:work_id) - [id])
  end

  def local_title
    return title if I18n.locale == :ja
    return title_en if title_en.present?

    title
  end

  def formatted_number(raw_number)
    return unless number_format

    number = raw_number.to_i
    return number_format.data[number - 1] if number_format.format.blank?

    number_format.format % number
  end

  def update_watchers_count!(prev_status_kind, next_status_kind)
    is_prev_positive = prev_status_kind.in?(Status::POSITIVE_KINDS)
    is_next_positive = next_status_kind.in?(Status::POSITIVE_KINDS)

    return if is_prev_positive && is_next_positive

    decrement!(:watchers_count) if is_prev_positive
    increment!(:watchers_count) if is_next_positive
  end
end
