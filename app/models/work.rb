# == Schema Information
#
# Table name: works
#
#  id                :integer          not null, primary key
#  season_id         :integer
#  sc_tid            :integer
#  title             :string(510)      not null
#  media             :integer          not null
#  official_site_url :string(510)      default(""), not null
#  wikipedia_url     :string(510)      default(""), not null
#  episodes_count    :integer          default(0), not null
#  watchers_count    :integer          default(0), not null
#  released_at       :date
#  created_at        :datetime
#  updated_at        :datetime
#  fetch_syobocal    :boolean          default(FALSE), not null
#  twitter_username  :string(510)
#  twitter_hashtag   :string(510)
#  released_at_about :string
#  items_count       :integer          default(0), not null
#
# Indexes
#
#  index_works_on_items_count  (items_count)
#  works_sc_tid_key            (sc_tid) UNIQUE
#  works_season_id_idx         (season_id)
#

class Work < ActiveRecord::Base
  extend Enumerize

  enumerize :media, in: { tv: 1, ova: 2, movie: 3, web: 4, other: 0 }
  has_paper_trail

  belongs_to :season
  has_many   :activities, foreign_key: :recipient_id, foreign_type: :recipient, dependent: :destroy
  has_many   :checkins, dependent: :destroy
  has_many   :checks,   dependent: :destroy
  has_many   :episodes, dependent: :destroy
  has_many   :items,    dependent: :destroy
  has_many   :programs, dependent: :destroy
  has_many   :statuses, dependent: :destroy

  validates :sc_tid, numericality: { only_integer: true }, uniqueness: true,
                     allow_blank: true
  validates :title, presence: true, uniqueness: true
  validates :media, presence: true
  validates :official_site_url, url: { allow_blank: true }
  validates :wikipedia_url, url: { allow_blank: true }

  scope :by_season, -> (season_slug) {
    return self if season_slug.blank?

    season = Season.find_by(slug: season_slug)

    return none if season.blank?

    where(season_id: season.id)
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


  # 作品のエピソード数分の空白文字列が入った配列を返す
  # Chart.jsのx軸のラベルを消すにはこれしか方法がなかったんだ…! たぶん…。
  def chart_labels
    episodes.pluck(:id).map {''}
  end

  def chart_values
    episodes.order(:sort_number).pluck(:checkins_count)
  end

  def checkins_count
    chart_values.reduce(&:+).presence || 0
  end

  def main_item
    items.find_by(main: true).presence || Item.find(1)
  end

  def comments_count
    episode_ids = episodes.pluck(:id)
    checkins = Checkin.where(episode_id: episode_ids).where('comment != ?', '')

    checkins.count
  end

  def twitter_url
    twitter_username.present? ? "https://twitter.com/#{twitter_username}" : ''
  end

  def twitter_hashtag_url
    twitter_hashtag.present? ? URI.encode("https://twitter.com/search?q=#{twitter_hashtag}&src=hash") : ''
  end

  def syobocal_url
    "http://cal.syoboi.jp/tid/#{sc_tid}"
  end

  def channels
    if episodes.present?
      programs = Program.where(episode_id: episodes.order(:sort_number).first.id)
      Channel.where(id: programs.pluck(:channel_id).uniq) if programs.present?
    end
  end

  def release_date
    released_at.presence || released_at_about.presence || ''
  end

  def current_season?
    season.present? && season.slug == ENV["ANNICT_CURRENT_SEASON"]
  end

  def next_season?
    season.present? && season.slug == ENV["ANNICT_NEXT_SEASON"]
  end
end
