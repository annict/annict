# frozen_string_literal: true
# == Schema Information
#
# Table name: records
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer
#  comment              :text
#  modify_comment       :boolean          default(FALSE), not null
#  twitter_url_hash     :string(510)
#  facebook_url_hash    :string(510)
#  twitter_click_count  :integer          default(0), not null
#  facebook_click_count :integer          default(0), not null
#  comments_count       :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  shared_twitter       :boolean          default(FALSE), not null
#  shared_facebook      :boolean          default(FALSE), not null
#  work_id              :integer          not null
#  rating               :float
#  multiple_record_id   :integer
#  oauth_application_id :integer
#  rating_state         :string
#  review_id            :integer
#  aasm_state           :string           default("published"), not null
#  locale               :string           default("other"), not null
#  impressions_count    :integer          default(0), not null
#
# Indexes
#
#  checkins_episode_id_idx                (episode_id)
#  checkins_facebook_url_hash_key         (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key          (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                   (user_id)
#  index_records_on_impressions_count     (impressions_count)
#  index_records_on_locale                (locale)
#  index_records_on_multiple_record_id    (multiple_record_id)
#  index_records_on_oauth_application_id  (oauth_application_id)
#  index_records_on_rating_state          (rating_state)
#  index_records_on_review_id             (review_id)
#  index_records_on_work_id               (work_id)
#

class Record < ApplicationRecord
  extend Enumerize
  include AASM
  include LocaleDetectable

  enumerize :rating_state, in: %i(bad average good great), scope: true

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :review, optional: true
  belongs_to :work
  belongs_to :episode, counter_cache: true, optional: true
  belongs_to :multiple_record, optional: true
  belongs_to :user, counter_cache: true
  has_many :comments, dependent: :destroy
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :comment, length: { maximum: 1_500 }
  validates :rating,
    allow_blank: true,
    numericality: {
      greater_than_or_equal_to: 1,
      less_than_or_equal_to: 5
    }

  scope :with_comment, -> { where.not(comment: ["", nil]) }
  scope :with_no_comment, -> { where(comment: ["", nil]) }

  after_destroy :expire_cache
  after_save :expire_cache

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

  def setup_shared_sns(user)
    self.shared_twitter =
      user.twitter.present? &&
      user.authorized_to?(:twitter, shareable: true) &&
      user.setting.share_record_to_twitter?
    self.shared_facebook =
      user.facebook.present? &&
      user.authorized_to?(:facebook, shareable: true) &&
      user.setting.share_record_to_facebook?
  end

  def shared_sns?
    twitter_url_hash.present? || facebook_url_hash.present? ||
      shared_twitter? || shared_facebook?
  end

  def update_share_record_status
    if user.setting.share_record_to_twitter? != shared_twitter?
      user.setting.update_column(:share_record_to_twitter, shared_twitter?)
    end

    if user.setting.share_record_to_facebook? != shared_facebook?
      user.setting.update_column(:share_record_to_facebook, shared_facebook?)
    end
  end

  def share_to_sns
    TwitterShareJob.perform_later(user_id, id) if shared_twitter?
    FacebookShareJob.perform_later(user_id, id) if shared_facebook?
  end

  private

  def expire_cache
    user.channel_works.find_by(work: work)&.touch
    user.touch(:record_cache_expired_at)
    activities.update_all(updated_at: Time.now)
  end
end
