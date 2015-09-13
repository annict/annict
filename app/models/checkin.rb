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
#  work_id              :integer
#
# Indexes
#
#  checkins_episode_id_idx         (episode_id)
#  checkins_facebook_url_hash_key  (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key   (twitter_url_hash) UNIQUE
#  checkins_user_id_idx            (user_id)
#  index_checkins_on_work_id       (work_id)
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
    if shared_twitter? || shared_facebook?
      user.update_column(:share_checkin, true) unless user.share_checkin?
    else
      user.update_column(:share_checkin, false) if user.share_checkin?
    end
  end

  def share_to_sns
    if shared_twitter?
      TwitterService.new(user).delay.share!(self)
    elsif shared_facebook?
      FacebookService.new(user).delay.share!(self)
    end
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
