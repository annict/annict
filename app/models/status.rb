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

class Status < ActiveRecord::Base
  include StatusCommon

  belongs_to :oauth_application, class_name: "Doorkeeper::Application"
  belongs_to :user, touch: true

  after_create :save_activity
  after_create :refresh_watchers_count
  after_create :update_channel_work
  after_create :finish_tips
  after_create :save_latest_status

  def self.initial
    order(:id).first
  end

  def self.initial?(status)
    count == 1 && initial.id == status.id
  end

  private

  def save_activity
    Activity.create do |a|
      a.user      = user
      a.recipient = work
      a.trackable = self
      a.action    = "create_status"
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
    watches  = [:wanna_watch, :watching, :watched]

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
    when 'wanna_watch', 'watching'
      ChannelWorkService.new(user).create(work)
    else
      ChannelWorkService.new(user).delete(work)
    end
  end

  private

  def finish_tips
    if user.statuses.initial?(self)
      UserTipsService.new(user).finish!(:status)
    end
  end

  def save_latest_status
    latest_status = user.latest_statuses.find_or_initialize_by(work: work)
    latest_status.kind = kind
    latest_status.watched_episode_ids = [] if %w(watched stop_watching).include?(kind)
    latest_status.save!
  end
end
