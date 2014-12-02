# == Schema Information
#
# Table name: checkins
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer          not null
#  comment              :text
#  spoil                :boolean          default(FALSE), not null
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
  attr_accessor :request_from_sns

  belongs_to :work
  belongs_to :episode,  counter_cache: true
  belongs_to :user,     counter_cache: true
  has_many   :comments, dependent: :destroy

  validates :comment, length: { maximum: 500 }

  before_update :check_comment_modified
  after_create  :save_activity
  after_create  :finish_tips
  after_destroy :delete_activity
  after_save    :update_share_checkin_status
  after_commit  :share_to_twitter
  after_commit  :share_to_facebook
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
    case user.provider_name
    when 'Twitter'
      self.shared_twitter = user.share_checkin?
    when 'Facebook'
      self.shared_facebook = user.share_checkin?
    end
  end

  def shared_sns?
    twitter_url_hash.present? || facebook_url_hash.present? ||
    shared_twitter? || shared_facebook?
  end


  private

  def share_to_twitter
    TwitterCheckinShareWorker.perform_async(id) if shared_twitter? && !request_from_sns
  end

  def share_to_facebook
    FacebookCheckinShareWorker.perform_async(id) if shared_facebook? && !request_from_sns
  end

  def save_activity
    Activity.create do |a|
      a.user      = user
      a.recipient = episode
      a.trackable = self
      a.action    = 'checkins.create'
    end
  end

  def delete_activity
    activity = Activity.find_by(trackable_id: id, trackable_type: 'Checkin')
    activity.destroy
  end

  def check_comment_modified
    self.modify_comment = true if comment_changed?
  end

  def update_share_checkin_status
    if !request_from_sns
      if shared_twitter? || shared_facebook?
        user.update_column(:share_checkin, true) unless user.share_checkin?
      else
        user.update_column(:share_checkin, false) if user.share_checkin?
      end
    end
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
end
