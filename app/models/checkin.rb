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
#
# Indexes
#
#  checkins_episode_id_idx         (episode_id)
#  checkins_facebook_url_hash_key  (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key   (twitter_url_hash) UNIQUE
#  checkins_user_id_idx            (user_id)
#

class Checkin < ActiveRecord::Base
  attr_accessor :facebook_share, :twitter_share

  belongs_to :episode,  counter_cache: true
  belongs_to :user,     counter_cache: true
  has_many   :comments, dependent: :destroy

  validates :comment, length: { maximum: 500 }

  after_create  :save_activity
  after_destroy :delete_activity
  after_save    :share_to_twitter
  after_save    :share_to_facebook
  before_update :check_comment_modified


  def generate_url_hash
    SecureRandom.urlsafe_base64.slice(0, 10)
  end

  def work
    episode.work
  end

  private

  def share_to_twitter
    TwitterCheckinShareWorker.perform_async(id) if twitter_share.present?
  end

  def share_to_facebook
    FacebookCheckinShareWorker.perform_async(id) if facebook_share.present?
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
end
