# frozen_string_literal: true
# == Schema Information
#
# Table name: checkins
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer          not null
#  comment              :text
#  twitter_url_hash     :string
#  twitter_click_count  :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  facebook_url_hash    :string
#  facebook_click_count :integer          default(0), not null
#  comments_count       :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  modify_comment       :boolean          default(FALSE), not null
#  shared_twitter       :boolean          default(FALSE), not null
#  shared_facebook      :boolean          default(FALSE), not null
#  work_id              :integer
#  rating               :float
#  multiple_record_id   :integer
#
# Indexes
#
#  index_checkins_on_facebook_url_hash   (facebook_url_hash) UNIQUE
#  index_checkins_on_multiple_record_id  (multiple_record_id)
#  index_checkins_on_twitter_url_hash    (twitter_url_hash) UNIQUE
#  index_checkins_on_work_id             (work_id)
#

class Checkin < ActiveRecord::Base
  belongs_to :work
  belongs_to :episode, counter_cache: true
  belongs_to :multiple_record
  belongs_to :user, counter_cache: true
  has_many   :comments,   dependent: :destroy
  has_many   :activities, dependent: :destroy, foreign_key: :trackable_id, foreign_type: :trackable
  has_many   :likes,      dependent: :destroy, foreign_key: :recipient_id, foreign_type: :recipient

  validates :comment, length: { maximum: 500 }
  validates :rating, numericality: {
    greater_than_or_equal_to: 0,
    less_than_or_equal_to: 5
  }

  scope :with_comment, -> { where.not(comment: '') }

  before_update :check_comment_modified
  before_save :update_rating
  after_create  :save_activity
  after_create  :finish_tips
  after_create  :update_latest_status

  def self.initial?(checkin)
    count == 1 && first.id == checkin.id
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

  def set_shared_sns(user)
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

  def share_to_sns(controller)
    TwitterService.new(user).delay.share!(self) if shared_twitter?
    FacebookService.new(user).share!(self, controller) if shared_facebook?
  end

  private

  def save_activity
    if multiple_record.blank?
      Activity.create do |a|
        a.user      = user
        a.recipient = episode
        a.trackable = self
        a.action    = "create_record"
      end
    end
  end

  def check_comment_modified
    self.modify_comment = true if comment_changed?
  end

  def finish_tips
    if user.checkins.initial?(self)
      UserTipsService.new(user).finish!(:checkin)
    end
  end

  def update_latest_status
    latest_status = user.latest_statuses.find_by(work: work)
    latest_status.append_episode(episode) if latest_status.present?
  end

  def update_rating
    self.rating = nil if rating.present? && rating < 1
  end
end
