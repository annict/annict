# frozen_string_literal: true
# == Schema Information
#
# Table name: episode_records
#
#  id                   :bigint           not null, primary key
#  body                 :text
#  comments_count       :integer          default(0), not null
#  deleted_at           :datetime
#  facebook_click_count :integer          default(0), not null
#  facebook_url_hash    :string(510)
#  likes_count          :integer          default(0), not null
#  locale               :string           default("other"), not null
#  modify_body          :boolean          default(FALSE), not null
#  rating               :float
#  rating_state         :string
#  twitter_click_count  :integer          default(0), not null
#  twitter_url_hash     :string(510)
#  created_at           :datetime
#  updated_at           :datetime
#  anime_id             :bigint           not null
#  episode_id           :bigint           not null
#  oauth_application_id :bigint
#  record_id            :bigint           not null
#  user_id              :bigint           not null
#
# Indexes
#
#  checkins_facebook_url_hash_key                       (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key                        (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                                 (user_id)
#  index_episode_records_on_anime_id                    (anime_id)
#  index_episode_records_on_episode_id_and_deleted_at   (episode_id,deleted_at)
#  index_episode_records_on_locale                      (locale)
#  index_episode_records_on_multiple_episode_record_id  (multiple_episode_record_id)
#  index_episode_records_on_oauth_application_id        (oauth_application_id)
#  index_episode_records_on_rating_state                (rating_state)
#  index_episode_records_on_record_id                   (record_id) UNIQUE
#  index_episode_records_on_review_id                   (review_id)
#
# Foreign Keys
#
#  checkins_episode_id_fk  (episode_id => episodes.id) ON DELETE => cascade
#  checkins_user_id_fk     (user_id => users.id) ON DELETE => cascade
#  checkins_work_id_fk     (anime_id => animes.id)
#  fk_rails_...            (multiple_episode_record_id => multiple_episode_records.id)
#  fk_rails_...            (oauth_application_id => oauth_applications.id)
#  fk_rails_...            (record_id => records.id)
#  fk_rails_...            (review_id => anime_records.id)
#

class EpisodeRecord < ApplicationRecord
  extend Enumerize

  include Localizable
  include Shareable
  include SoftDeletable

  self.ignored_columns = %w(aasm_state multiple_episode_record_id review_id shared_facebook shared_twitter)

  enumerize :rating_state, in: Record::RATING_STATES, scope: true

  counter_culture :episode
  counter_culture :episode, column_name: -> (episode_record) { episode_record.body.present? ? :episode_record_bodies_count : nil }
  counter_culture :user

  attr_accessor :share_to_twitter, :mutation_error

  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :record
  belongs_to :work
  belongs_to :episode
  belongs_to :multiple_episode_record, optional: true
  belongs_to :user
  has_many :comments, dependent: :destroy
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :body, length: { maximum: 1_048_596 }
  validates :rating,
    allow_blank: true,
    numericality: {
      greater_than_or_equal_to: 1,
      less_than_or_equal_to: 5
    }

  scope :with_body, -> { where.not(body: ["", nil]) }
  scope :with_no_body, -> { where(body: ["", nil]) }

  def self.initial?(record)
    count == 1 && first.id == record.id
  end

  def self.rating_state_order(direction = :asc)
    direction = direction.in?(%i(asc desc)) ? direction : :asc
    order <<-SQL
    CASE
      WHEN rating_state = 'bad' THEN '0'
      WHEN rating_state = 'average' THEN '1'
      WHEN rating_state = 'good' THEN '2'
      WHEN rating_state = 'great' THEN '3'
    END #{direction.upcase} NULLS LAST
    SQL
  end

  def rating=(value)
    return super if value.to_f.between?(1, 5)
    write_attribute :rating, nil
  end

  def rating_to_rating_state
    case rating
    when 1.0...3.0 then :bad
    when 3.0...3.5 then :average
    when 3.5...4.5 then :good
    when 4.5..5.0 then :great
    end
  end

  def initial
    order(:id).first
  end

  def generate_url_hash
    SecureRandom.urlsafe_base64.slice(0, 10)
  end

  def share_url
    "#{user.preferred_annict_url}/@#{user.username}/records/#{record.id}"
  end

  def facebook_share_title
    "#{work.title} #{episode.title_with_number}"
  end

  def twitter_share_body
    work_title = work.local_title
    title = self.body.present? ? work_title.truncate(30) : work_title
    comment = self.body.present? ? "#{self.body} / " : ""
    episode_number = episode.local_number
    share_url = share_url_with_query(:twitter)
    share_hashtag = work.hashtag_with_hash

    base_body = if user.locale == "ja"
      "%s#{title} #{episode_number} を見ました #{share_url} #{share_hashtag}"
    else
      "%sWatched: #{title} #{episode_number} #{share_url} #{share_hashtag}"
    end

    body = base_body % comment
    body_without_url = body.sub(share_url, "")
    return body if body_without_url.length <= 130

    comment = comment.truncate(comment.length - (body_without_url.length - 130)) + " / "
    base_body % comment
  end

  def facebook_share_body
    return self.body if self.body.present?

    if user.locale == "ja"
      "見ました。"
    else
      "Watched."
    end
  end

  def needs_single_activity_group?
    body.present?
  end
end
