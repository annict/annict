# == Schema Information
#
# Table name: statuses
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  work_id     :integer          not null
#  kind        :integer          not null
#  latest      :boolean          default(FALSE), not null
#  likes_count :integer          default(0), not null
#  created_at  :datetime
#  updated_at  :datetime
#
# Indexes
#
#  statuses_user_id_idx  (user_id)
#  statuses_work_id_idx  (work_id)
#

class Status < ActiveRecord::Base
  extend Enumerize

  enumerize :kind, in: { wanna_watch: 1, watching: 2, watched: 3, on_hold: 5, stop_watching: 4 }, scope: true

  belongs_to :user
  belongs_to :work

  scope :latest, -> { where(latest: true) }
  scope :watching, -> { latest.with_kind(:watching) }
  scope :wanna_watch_and_watching, -> { latest.with_kind(:wanna_watch, :watching) }

  after_create :change_latest
  after_create :save_activity
  after_create :refresh_watchers_count
  after_create :update_recommendable
  after_create :update_channel_work
  after_create :finish_tips
  after_commit :publish_events, on: :create


  def self.initial
    order(:id).first
  end

  def self.initial?(status)
    self.count == 1 && initial.id == status.id
  end

  def self.kind_of(work)
    latest.find_by(work_id: work.id)
  end


  private

  def change_latest
    latest_status = user.statuses.find_by(work_id: work.id, latest: true)
    latest_status.update_column(:latest, false) if latest_status.present?

    update_column(:latest, true)
  end

  def save_activity
    Activity.create do |a|
      a.user      = user
      a.recipient = work
      a.trackable = self
      a.action    = 'statuses.create'
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
    statuses = user.statuses.where(work_id: work.id).includes(:work).last(2)
    prev_status = statuses.first.kind.to_sym
    new_status  = statuses.last.kind.to_sym

    if statuses.length == 2
      return :watch if !watches.include?(prev_status) && watches.include?(new_status)
      return :drop  if watches.include?(prev_status)  && !watches.include?(new_status)
      return :keep # 見たい系 -> 見たい系 または 中止系 -> 中止系
    end

    watches.include?(new_status) ? :watch : :drop_first
  end

  # ステータスの変更があったとき、「Recommendable」の `like`, `dislike` などを呼び出して
  # オススメ作品を更新する
  def update_recommendable
    if become_to == :watch
      user.undislike(work) if user.dislikes?(work)
      user.like(work)
    elsif become_to == :drop
      user.unlike(work) if user.likes?(work)
      user.dislike(work)
    elsif become_to == :drop_first
      user.dislike(work)
    end
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

  def publish_events
    FirstStatusesEvent.publish(:create, self) if user.statuses.initial?(self)
    StatusesEvent.publish(:create, self)
  end

  def finish_tips
    if user.statuses.initial?(self)
      tip = Tip.find_by(partial_name: 'status')
      user.tips.finish!(tip)
    end
  end
end
