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
#
# Indexes
#
#  index_checkins_on_facebook_url_hash  (facebook_url_hash) UNIQUE
#  index_checkins_on_twitter_url_hash   (twitter_url_hash) UNIQUE
#  index_checkins_on_work_id            (work_id)
#

class Checkin < ActiveRecord::Base
  belongs_to :work
  belongs_to :episode, counter_cache: true
  belongs_to :user,    counter_cache: true
  has_many   :comments,   dependent: :destroy
  has_many   :activities, dependent: :destroy, foreign_key: :trackable_id, foreign_type: :trackable
  has_many   :likes,      dependent: :destroy, foreign_key: :recipient_id, foreign_type: :recipient

  validates :comment, length: { maximum: 500 }

  scope :with_comment, -> { where.not(comment: '') }

  before_update :check_comment_modified
  after_create  :save_activity
  after_create  :finish_tips
  after_create  :refresh_check
  after_commit  :publish_events, on: :create


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

  def share_to_sns
    TwitterService.new(user).delay.share!(self) if shared_twitter?
    FacebookService.new(user).delay.share!(self) if shared_facebook?
  end

  private

  def save_activity
    Activity.create do |a|
      a.user      = user
      a.recipient = episode
      a.trackable = self
      a.action    = 'checkins.create'
    end
  end

  def check_comment_modified
    self.modify_comment = true if comment_changed?
  end

  def publish_events
    FirstCheckinsEvent.publish(:create, self) if user.checkins.initial?(self)
    CheckinsEvent.publish(:create, self)
  end

  def finish_tips
    if user.checkins.initial?(self)
      UserTipsService.new(user).finish!(:checkin)
    end
  end

  def refresh_check
    check = user.checks.find_by(work_id: work.id)
    check.update_episode_to_next(checkin: self) if check.present?
  end
end
