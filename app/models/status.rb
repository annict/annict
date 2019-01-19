# frozen_string_literal: true

# == Schema Information
#
# Table name: statuses
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  work_id              :integer          not null
#  kind                 :integer          not null
#  likes_count          :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  oauth_application_id :integer
#
# Indexes
#
#  index_statuses_on_oauth_application_id  (oauth_application_id)
#  statuses_user_id_idx                    (user_id)
#  statuses_work_id_idx                    (work_id)
#

class Status < ApplicationRecord
  include StatusCommon
  include Shareable

  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  after_create :finish_tips
  after_create :refresh_watchers_count
  after_create :save_activity
  after_create :save_latest_status
  after_create :update_channel_work
  after_destroy :expire_cache
  after_save :expire_cache

  def self.initial
    order(:id).first
  end

  def self.initial?(status)
    count == 1 && initial.id == status.id
  end

  def share_to_sns
    ShareStatusToTwitterJob.perform_later(user.id, id) if user.setting.share_status_to_twitter? && user.authorized_to?(:twitter, shareable: true)
  end

  # Do not use helper methods via Draper when the method is used in ActiveJob
  # https://github.com/drapergem/draper/issues/655
  def share_url
    "#{user.annict_url}/@#{user.username}/#{kind}"
  end

  def facebook_share_title
    work.local_title
  end

  def twitter_share_body
    work_title = work.local_title
    share_url = share_url_with_query(:twitter)

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました #{share_url}"
    else
      "Changed %s's status to \"#{kind_text}\". Anime list: #{share_url}"
    end

    base_body % work_title
  end

  def facebook_share_body
    work_title = work.local_title

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました。"
    else
      "Changed %s's status to \"#{kind_text}\"."
    end

    base_body % work_title
  end

  private

  def save_activity
    Activity.create do |a|
      a.user = user
      a.recipient = work
      a.trackable = self
      a.action = "create_status"
      a.work = work
      a.status = self
    end
  end

  def refresh_watchers_count
    if become_to == :watch
      work.increment!(:watchers_count)
    elsif become_to == :drop
      work.decrement!(:watchers_count)
    end
  end

  # ステータスを何から何に変えたかを返す
  def become_to
    watches = %i(wanna_watch watching watched)

    if last_2_statuses.length == 2
      return :watch if !watches.include?(prev_status) && watches.include?(new_status)
      return :drop  if watches.include?(prev_status)  && !watches.include?(new_status)
      return :keep # 見たい系 -> 見たい系 または 中止系 -> 中止系
    end

    watches.include?(new_status) ? :watch : :drop_first
  end

  def last_2_statuses
    @last_2_statuses ||= user.statuses.where(work_id: work.id).includes(:work).last(2)
  end

  def prev_status
    @prev_status ||= last_2_statuses.first.kind.to_sym
  end

  def new_status
    @new_status ||= last_2_statuses.last.kind.to_sym
  end

  def update_channel_work
    case kind
    when "wanna_watch", "watching"
      ChannelWorkService.new(user).create(work)
    else
      ChannelWorkService.new(user).delete(work)
    end
  end

  def finish_tips
    UserTipsService.new(user).finish!(:status) if user.statuses.initial?(self)
  end

  def save_latest_status
    latest_status = user.latest_statuses.find_or_initialize_by(work: work)
    latest_status.kind = kind
    latest_status.watched_episode_ids = [] if %w(watched stop_watching).include?(kind)
    latest_status.save!
  end

  def expire_cache
    user.touch(:status_cache_expired_at)
  end
end
