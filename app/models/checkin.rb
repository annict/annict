# == Schema Information
#
# Table name: checkins
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer          not null
#  comment              :text
#  twitter_url_hash     :string(255)
#  twitter_click_count  :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  spoil                :boolean          default(FALSE), not null
#  facebook_url_hash    :string(255)
#  facebook_click_count :integer          default(0), not null
#  comments_count       :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  modify_comment       :boolean          default(FALSE), not null
#  shared_twitter       :boolean          default(FALSE), not null
#  shared_facebook      :boolean          default(FALSE), not null
#
# Indexes
#
#  index_checkins_on_facebook_url_hash  (facebook_url_hash) UNIQUE
#  index_checkins_on_twitter_url_hash   (twitter_url_hash) UNIQUE
#

class Checkin < ActiveRecord::Base
  attr_accessor :request_from_sns

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
    FirstCheckinsEvent.publish(:create, self) if user.first_checkin?(self)
    CheckinsEvent.publish(:create, self)
  end

  def finish_tips
    if user.first_checkin?(self)
      tip = Tip.find_by(partial_name: 'checkin')
      user.finish_tip!(tip)
    end
  end
end
