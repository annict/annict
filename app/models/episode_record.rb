# frozen_string_literal: true
# == Schema Information
#
# Table name: episode_records
#
#  id                         :integer          not null, primary key
#  aasm_state                 :string           default("published"), not null
#  body                       :text
#  comments_count             :integer          default(0), not null
#  deleted_at                 :datetime
#  facebook_click_count       :integer          default(0), not null
#  facebook_url_hash          :string(510)
#  likes_count                :integer          default(0), not null
#  locale                     :string           default("other"), not null
#  modify_body                :boolean          default(FALSE), not null
#  rating                     :float
#  rating_state               :string
#  shared_facebook            :boolean          default(FALSE), not null
#  shared_twitter             :boolean          default(FALSE), not null
#  twitter_click_count        :integer          default(0), not null
#  twitter_url_hash           :string(510)
#  created_at                 :datetime
#  updated_at                 :datetime
#  episode_id                 :integer          not null
#  multiple_episode_record_id :integer
#  oauth_application_id       :integer
#  record_id                  :integer          not null
#  review_id                  :integer
#  user_id                    :integer          not null
#  work_id                    :integer          not null
#
# Indexes
#
#  checkins_episode_id_idx                              (episode_id)
#  checkins_facebook_url_hash_key                       (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key                        (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                                 (user_id)
#  index_episode_records_on_deleted_at                  (deleted_at)
#  index_episode_records_on_locale                      (locale)
#  index_episode_records_on_multiple_episode_record_id  (multiple_episode_record_id)
#  index_episode_records_on_oauth_application_id        (oauth_application_id)
#  index_episode_records_on_rating_state                (rating_state)
#  index_episode_records_on_record_id                   (record_id) UNIQUE
#  index_episode_records_on_review_id                   (review_id)
#  index_episode_records_on_work_id                     (work_id)
#
# Foreign Keys
#
#  checkins_episode_id_fk  (episode_id => episodes.id) ON DELETE => cascade
#  checkins_user_id_fk     (user_id => users.id) ON DELETE => cascade
#  checkins_work_id_fk     (work_id => works.id)
#  fk_rails_...            (multiple_episode_record_id => multiple_episode_records.id)
#  fk_rails_...            (oauth_application_id => oauth_applications.id)
#  fk_rails_...            (record_id => records.id)
#  fk_rails_...            (review_id => work_records.id)
#

class EpisodeRecord < ApplicationRecord
  extend Enumerize

  include Localizable
  include Shareable
  include SoftDeletable

  enumerize :rating_state, in: Record::RATING_STATES, scope: true

  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :record
  belongs_to :review, optional: true
  belongs_to :work
  belongs_to :episode, counter_cache: true
  belongs_to :multiple_episode_record, optional: true
  belongs_to :user, counter_cache: true
  has_many :comments, dependent: :destroy
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :body, length: { maximum: 1000 }
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

  def work
    episode.work
  end

  def setup_shared_sns(user)
    self.shared_twitter =
      user.twitter.present? &&
      user.authorized_to?(:twitter, shareable: true) &&
      user.setting.share_record_to_twitter?
  end

  def shared_sns?
    twitter_url_hash.present? || shared_twitter?
  end

  def update_share_record_status
    if user.setting.share_record_to_twitter? != shared_twitter?
      user.setting.update_column(:share_record_to_twitter, shared_twitter?)
    end
  end

  def share_to_sns
    ShareEpisodeRecordToTwitterJob.perform_later(user_id, id) if shared_twitter?
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
end
