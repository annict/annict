# frozen_string_literal: true
# == Schema Information
#
# Table name: works
#
#  id                           :integer          not null, primary key
#  season_id                    :integer
#  sc_tid                       :integer
#  title                        :string(510)      not null
#  media                        :integer          not null
#  official_site_url            :string(510)      default(""), not null
#  wikipedia_url                :string(510)      default(""), not null
#  auto_episodes_count          :integer          default(0), not null
#  watchers_count               :integer          default(0), not null
#  released_at                  :date
#  created_at                   :datetime
#  updated_at                   :datetime
#  twitter_username             :string(510)
#  twitter_hashtag              :string(510)
#  released_at_about            :string
#  aasm_state                   :string           default("published"), not null
#  number_format_id             :integer
#  title_kana                   :string           default(""), not null
#  title_ro                     :string           default(""), not null
#  title_en                     :string           default(""), not null
#  official_site_url_en         :string           default(""), not null
#  wikipedia_url_en             :string           default(""), not null
#  synopsis                     :text             default(""), not null
#  synopsis_en                  :text             default(""), not null
#  synopsis_source              :string           default(""), not null
#  synopsis_source_en           :string           default(""), not null
#  mal_anime_id                 :integer
#  facebook_og_image_url        :string           default(""), not null
#  twitter_image_url            :string           default(""), not null
#  recommended_image_url        :string           default(""), not null
#  season_year                  :integer
#  season_name                  :integer
#  key_pv_id                    :integer
#  manual_episodes_count        :integer
#  no_episodes                  :boolean          default(FALSE), not null
#  work_records_count           :integer          default(0), not null
#  started_on                   :date
#  ended_on                     :date
#  score                        :float
#  ratings_count                :integer          default(0), not null
#  satisfaction_rate            :float
#  records_count                :integer          default(0), not null
#  work_records_with_body_count :integer          default(0), not null
#
# Indexes
#
#  index_works_on_aasm_state                           (aasm_state)
#  index_works_on_key_pv_id                            (key_pv_id)
#  index_works_on_number_format_id                     (number_format_id)
#  index_works_on_ratings_count                        (ratings_count)
#  index_works_on_satisfaction_rate                    (satisfaction_rate)
#  index_works_on_satisfaction_rate_and_ratings_count  (satisfaction_rate,ratings_count)
#  index_works_on_score                                (score)
#  index_works_on_season_year                          (season_year)
#  index_works_on_season_year_and_season_name          (season_year,season_name)
#  works_sc_tid_key                                    (sc_tid) UNIQUE
#  works_season_id_idx                                 (season_id)
#

class Work < ApplicationRecord
  extend Enumerize
  include AASM
  include DbActivityMethods
  include Itemable
  include RootResourceCommon

  DIFF_FIELDS = %i(
    sc_tid title title_kana title_en media official_site_url
    official_site_url_en wikipedia_url wikipedia_url_en twitter_username
    twitter_hashtag number_format_id synopsis synopsis_en synopsis_source
    synopsis_source_en mal_anime_id season_year season_name manual_episodes_count
    started_on ended_on
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
  has_many :casts, dependent: :destroy
  has_many :program_details, dependent: :destroy
  has_many :resource_items, dependent: :destroy, class_name: "WorkItem"
  has_many :series_works, dependent: :destroy
  has_many :staffs, dependent: :destroy
  has_many :work_taggings, dependent: :destroy
  has_many :activities,
    foreign_key: :recipient_id,
    foreign_type: :recipient,
    dependent: :destroy
  has_many :cast_people, through: :casts, source: :person
  has_many :channel_works, dependent: :destroy
  has_many :characters, through: :casts
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy
  has_many :episode_records, dependent: :destroy
  has_many :episodes, dependent: :destroy
  has_many :items, through: :resource_items
  has_many :latest_statuses, dependent: :destroy
  has_many :organizations,
    through: :staffs,
    source: :resource,
    source_type: "Organization"
  has_many :programs, dependent: :destroy
  has_many :pvs, dependent: :destroy
  has_many :records, dependent: :destroy
  has_many :series_list, through: :series_works, source: :series
  has_many :statuses, dependent: :destroy
  has_many :staff_people, through: :staffs, source: :resource, source_type: "Person"
  has_many :channels, through: :program_details
  has_many :work_records, dependent: :destroy
  has_many :work_tags, through: :work_taggings
  has_one :work_image, dependent: :destroy

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

  scope :by_no_season, -> {
    where(season_year: nil, season_name: nil)
  }

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

  scope :tracked_by, -> (user) {
    joins(
      "INNER JOIN (
        SELECT DISTINCT work_id, MAX(id) AS record_id FROM records
          WHERE records.user_id = #{user.id} GROUP BY work_id
      ) AS c2 ON works.id = c2.work_id"
    )
  }

  # 作品画像が設定されていない作品
  scope :image_not_attached, -> {
    joins("LEFT OUTER JOIN work_images ON work_images.work_id = works.id").
      where("work_images.id IS NULL")
  }

  scope :order_by_season, ->(type = :asc) {
    order(season_year: type, season_name: type)
  }

  scope :gt_current_season, -> {
    season = Season.find_by_slug(ENV.fetch("ANNICT_CURRENT_SEASON"))

    where("season_year >= ? AND season_name > ?", season.year, season.name_value).
      or(where("season_year > ?", season.year)).
      or(where(season_year: season.year, season_name: nil)).
      or(where(season_year: nil))
  }

  def self.statuses(work_ids, user)
    work_ids = work_ids.uniq
    latest_statuses = LatestStatus.where(user: user, work_id: work_ids)

    work_ids.map do |work_id|
      {
        work_id: work_id,
        kind: latest_statuses.select { |ls| ls.work_id == work_id }.first&.kind.presence || "no_select"
      }
    end
  end

  def self.work_tags_data(works, user)
    work_ids = works.pluck(:id)
    work_taggings = WorkTagging.where(user: user, work_id: work_ids)
    work_tags = WorkTag.where(id: work_taggings.pluck(:work_tag_id))

    work_ids.map do |work_id|
      work_tag_ids = work_taggings.
        select { |wt| wt.work_id == work_id }.
        map(&:work_tag_id)

      {
        work_id: work_id,
        work_tags: work_tags.select { |wt| wt.id.in?(work_tag_ids) }
      }
    end
  end

  def self.work_comment_data(works, user)
    work_ids = works.pluck(:id)
    work_comments = WorkComment.where(user: user, work_id: work_ids)

    work_ids.map do |work_id|
      {
        work_id: work_id,
        work_comment: work_comments.select { |c| c.work_id == work_id }.first
      }
    end
  end

  def self.watching_friends_data(work_ids, user)
    work_ids = work_ids.uniq
    status_kinds = %w(wanna_watch watching watched)
    users = user.followings.published.includes(:profile)
    user_ids = users.pluck(:id)
    latest_statuses = LatestStatus.
      where(work: work_ids, user: user_ids).
      with_kind(*status_kinds)

    work_ids.map do |work_id|
      latest_statuses_ = latest_statuses.select { |ls| ls.work_id == work_id }
      users_ = users.select { |u| u.id.in?(latest_statuses_.map(&:user_id)) }
      users_data = users_.map do |u|
        latest_status = latest_statuses_.select do |ls|
          ls.user_id == u.id && ls.work_id == work_id
        end.first

        {
          user: u,
          latest_status_id: latest_status.id
        }
      end

      {
        work_id: work_id,
        users_data: users_data
      }
    end
  end

  def self.pvs_data(works)
    work_ids = works.pluck(:id)
    pvs = Pv.published.where(work_id: work_ids)

    work_ids.map do |work_id|
      {
        work_id: work_id,
        pvs: pvs.select { |p| p.work_id == work_id }
      }
    end
  end

  def self.casts_data(works)
    work_ids = works.pluck(:id)
    casts = Cast.published.where(work_id: work_ids).includes(:person, :character)

    work_ids.map do |work_id|
      {
        work_id: work_id,
        casts: casts.select { |c| c.work_id == work_id }
      }
    end
  end

  def self.staffs_data(works, major: false)
    work_ids = works.pluck(:id)
    staffs = Staff.published.where(work_id: work_ids).includes(:resource)
    staffs = staffs.major if major

    work_ids.map do |work_id|
      {
        work_id: work_id,
        staffs: staffs.select { |s| s.work_id == work_id }
      }
    end
  end

  def self.program_details_data(works, only_vod: false)
    work_ids = works.pluck(:id)
    program_details = ProgramDetail.published.where(work_id: work_ids).includes(:channel)
    if only_vod
      program_details = program_details.
        joins(:channel).
        where(channels: { vod: true })
    end

    work_ids.map do |work_id|
      {
        work_id: work_id,
        program_details: program_details.select { |pd| pd.work_id == work_id }
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

  # 作品のエピソード数分の空白文字列が入った配列を返す
  # Chart.jsのx軸のラベルを消すにはこれしか方法がなかったんだ…! たぶん…。
  def chart_labels
    episodes.published.pluck(:id).map { "" }
  end

  def chart_values
    episodes.published.order(:sort_number).pluck(:episode_records_count)
  end

  def comments_count
    episode_ids = episodes.pluck(:id)
    records = Record.where(episode_id: episode_ids).where("comment != ?", "")

    records.count
  end

  def sync_with_syobocal?
    sc_tid.present?
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
    "https://twitter.com/#{twitter_username}"
  end

  def twitter_hashtag_url
    URI.encode("https://twitter.com/search?q=##{twitter_hashtag}&src=hash")
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

  def episodes_count
    manual_episodes_count.presence || auto_episodes_count
  end

  def hashtag_with_hash
    return "" if twitter_hashtag.blank?
    "##{twitter_hashtag}"
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

  def watchers_chart_dataset
    (3.months.ago.to_date..Date.today).map do |date|
      count = latest_statuses.with_kind(:wanna_watch, :watching, :watched).before(date).count
      {
        date: date.to_time.to_datetime.strftime("%Y/%m/%d"),
        value: count
      }
    end.to_json
  end

  def status_chart_dataset
    LatestStatus.kind.values.map do |kind|
      kind_count = latest_statuses.with_kind(kind).count
      {
        name: kind.text,
        value: kind_count
      }
    end.to_json
  end

  def image_color_rgb
    work_image&.color_rgb.presence || "255,255,255"
  end

  def image_text_color_rgb
    work_image&.text_color_rgb.presence || "0,0,0"
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
end
