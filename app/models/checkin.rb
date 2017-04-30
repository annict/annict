# frozen_string_literal: true
# == Schema Information
#
# Table name: checkins
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer          not null
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
#
# Indexes
#
#  checkins_episode_id_idx                 (episode_id)
#  checkins_facebook_url_hash_key          (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key           (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                    (user_id)
#  index_checkins_on_multiple_record_id    (multiple_record_id)
#  index_checkins_on_oauth_application_id  (oauth_application_id)
#  index_checkins_on_work_id               (work_id)
#

class Checkin < ActiveRecord::Base
  belongs_to :oauth_application, class_name: "Doorkeeper::Application"
  belongs_to :work
  belongs_to :episode, counter_cache: true
  belongs_to :multiple_record
  belongs_to :user, counter_cache: true
  has_many :comments, dependent: :destroy
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :comment, length: { maximum: 1000 }
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

  def self.initial?(checkin)
    count == 1 && first.id == checkin.id
  end

  def self.avg_rating
    ratings = pluck(:rating).select(&:present?)
    return if ratings.blank?
    (ratings.inject { |sum, rating| sum + rating } / ratings.count).round(1)
  end

  def rating=(value)
    return super if value.to_f.between?(1, 5)
    write_attribute :rating, nil
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
    self.shared_twitter = user.setting.share_record_to_twitter?
    self.shared_facebook = user.setting.share_record_to_facebook?
  end

  def shared_sns?
    twitter_url_hash.present? || facebook_url_hash.present? ||
      shared_twitter? || shared_facebook?
  end

  def update_share_checkin_status
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
  end
end
