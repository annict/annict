# == Schema Information
#
# Table name: users
#
#  id                   :integer          not null, primary key
#  username             :string(510)      not null
#  email                :string(510)      not null
#  role                 :integer          not null
#  encrypted_password   :string(510)      default(""), not null
#  remember_created_at  :datetime
#  sign_in_count        :integer          default(0), not null
#  current_sign_in_at   :datetime
#  last_sign_in_at      :datetime
#  current_sign_in_ip   :string(510)
#  last_sign_in_ip      :string(510)
#  confirmation_token   :string(510)
#  confirmed_at         :datetime
#  confirmation_sent_at :datetime
#  unconfirmed_email    :string(510)
#  checkins_count       :integer          default(0), not null
#  notifications_count  :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  share_checkin        :boolean          default(FALSE)
#
# Indexes
#
#  users_confirmation_token_key  (confirmation_token) UNIQUE
#  users_email_key               (email) UNIQUE
#  users_username_key            (username) UNIQUE
#

class User < ActiveRecord::Base
  # registrations#createが実行されたあとメールアドレスの確認を挟まず
  # ログインできるようにするため、Confirmableモジュールを直接includeする
  include Devise::Models::Confirmable

  extend Enumerize

  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :recoverable, :rememberable
  devise :database_authenticatable, :omniauthable, :registerable,
         :trackable, omniauth_providers: [:facebook, :twitter]

  enumerize :role, in: { user: 0, admin: 1, editor: 2 }, default: :user

  recommends :works

  has_many :activities,    dependent: :destroy
  has_many :channel_works, dependent: :destroy
  has_many :checkins,      dependent: :destroy
  has_many :follows,       dependent: :destroy
  has_many :followings,    through:   :follows
  has_many :r_likes,       dependent: :destroy, class_name: 'Like'
  has_many :notifications, dependent: :destroy
  has_many :providers,     dependent: :destroy
  has_many :receptions,    dependent: :destroy
  has_many :channels,      through:   :receptions
  has_many :statuses,      dependent: :destroy
  has_one  :profile,       dependent: :destroy

  validates :email, presence: true, uniqueness: true, email: true
  validates :username, presence: true, uniqueness: true, length: { maximum: 20 },
                       format: { with: /\A[A-Za-z0-9_]+\z/ }
  validates :terms, acceptance: true

  after_commit  :publish_events, on: :create


  def first_status
    statuses.order(:id).first
  end

  def first_status?(status)
    statuses.count == 1 && statuses.first.id == status.id
  end

  def latest_statuses
    statuses.where(latest: true)
  end

  def watching_statuses
    latest_statuses.with_kind(:watching)
  end

  def wanna_watch_or_watching_statuses
    latest_statuses.with_kind(:wanna_watch, :watching)
  end

  def status(work)
    latest_statuses.find_by(work_id: work.id)
  end

  def watching_works
    Work.joins(:statuses).merge(watching_statuses)
  end

  def wanna_watch_or_watching_works
    Work.joins(:statuses).merge(wanna_watch_or_watching_statuses)
  end

  def unknown_works
    Work.where.not(id: latest_statuses.pluck(:work_id))
  end

  # 指定した作品の中のチェックインしていないエピソードを返す
  def unchecked_episodes(work)
    episode_ids = work.episodes.pluck(:id)
    checked_episode_ids = checkins.pluck(:episode_id)
    work.episodes.where(id: (episode_ids - checked_episode_ids))
  end

  # チェックインしていないエピソードと紐づく番組情報を返す
  def unchecked_programs
    program_ids = []

    channel_works.includes(:work).where(work_id: wanna_watch_or_watching_works).references(:work).each do |cw|
      episode_ids = unchecked_episodes(cw.work).pluck(:id)
      program_ids << Program.where(channel_id: cw.channel_id, episode_id: episode_ids).pluck(:id)
    end

    Program.where(id: program_ids.flatten)
  end

  def build_relations(oauth)
    self.providers.build do |p|
      p.name             = oauth[:provider]
      p.uid              = oauth[:uid]
      p.token            = oauth[:credentials][:token]
      p.token_expires_at = oauth[:credentials][:expires_at]
      p.token_secret     = oauth[:credentials][:secret]
    end

    self.build_profile do |p|
      p.name        = oauth[:info][:name].presence || oauth[:info][:nickname]
      p.description = oauth[:info][:description]
      p.avatar_url  = get_large_avatar_image(oauth[:provider], oauth[:info][:image])
    end

    self
  end

  def following?(user)
    followings.where(id: user.id).present?
  end

  def followers
    Follow.where(following_id: id)
  end

  def follow(user)
    follows.create(following: user) unless following?(user)
  end

  def unfollow(user)
    following = follows.where(following_id: user.id).first
    following.destroy if following.present?
  end

  def receiving?(channel)
    receptions.where(channel_id: channel.id).present?
  end

  def receive(channel)
    unless receiving?(channel)
      receptions.create(channel: channel)
      ChannelWorksCreatingWorker.perform_async(id, channel.id)
    end
  end

  def unreceive(channel)
    reception = receptions.where(channel_id: channel.id).first

    if reception.present?
      reception.destroy
      ChannelWorksDestroyingWorker.perform_async(id, channel.id)
    end
  end

  def trim_username!
    # Facebookからのユーザ登録のとき `username` に「.」が含まれている可能性があるので除去する
    username.delete!('.')
  end

  def authorized_to?(provider)
    providers.pluck(:name).include?(provider.to_s)
  end

  def following_activities
    following_ids = followings.pluck(:id)
    following_ids << self.id

    Activity.where(user_id: following_ids)
  end

  def like_r?(recipient)
    r_likes.where(recipient: recipient).present?
  end

  #「Recommendable」の `like` メソッドと衝突したため、"_r" というサフィックスをつける羽目になった
  def like_r(recipient)
    r_likes.create(recipient: recipient) unless like_r?(recipient)
  end

  def unlike_r(recipient)
    like = r_likes.where(recipient: recipient).first

    like.destroy if like.present?
  end

  def watching_count
    statuses.where(latest: true).with_kind(:watching).count
  end

  def social_friends
    provider = providers.first
    uids = case provider.name
      when 'twitter'  then twitter_uids
      when 'facebook' then facebook_uids
      end

    User.joins(:providers).where(providers: { name: provider.name, uid: uids })
  end

  def provider_name
    providers.first.name.humanize
  end

  def works_on(status_kind)
    work_ids = latest_statuses.with_kind(status_kind).pluck(:work_id)

    Work.where(id: work_ids).order(released_at: :desc)
  end

  def read_notifications!
    transaction do
      unread_count = notifications.unread.update_all(read: true)
      decrement!(:notifications_count, unread_count)
    end
  end

  def background_image
    profile.background_image.presence || profile.avatar
  end

  def checkin_chart_labels
    today = Date.today
    week_days = (today - 6.days)..today

    week_days.map { |day| day.strftime('%-m/%d') }
  end

  def checkin_chart_values
    today = Date.today
    week_days = (today - 6.days).beginning_of_day..today.end_of_day
    weekly_checkins = checkins.where(created_at: week_days)
    weekly_checkins = weekly_checkins
                        .select('date(created_at) as checkins_day, count(*) as checkins_count')
                        .group('date(created_at)')

    weekly_checkins_hash = {}
    weekly_checkins.each do |checkin|
      weekly_checkins_hash[checkin.checkins_day.strftime('%-m/%d')] = checkin.checkins_count
    end

    checkins = []
    checkin_chart_labels.each do |label|
      checkins << (weekly_checkins_hash[label].presence || 0)
    end

    checkins
  end

  def fastest_channel(work)
    receivable_channel_ids = receptions.pluck(:channel_id)

    if receivable_channel_ids.present? && work.episodes.present?
      conditions = { channel_id: receivable_channel_ids, episode: work.episodes.first }
      fastest_program = Program.where(conditions).order(:started_at).first

      fastest_program.present? ? fastest_program.channel : nil
    end
  end

  # チャンネルと紐付いていない作品と最速放送チャンネルとを紐付ける
  def create_channel_work(work)
    channel_work = channel_works.find_by(work_id: work.id)

    if channel_work.blank?
      channel = fastest_channel(work)

      channel_works.create(work: work, channel: channel) if channel.present?
    end
  end

  # まだ番組情報が存在しなかったために `channel_works` に
  # どのチャンネルで見るかの情報が保存されなかった作品にcronで対処するためのメソッド
  def update_channel_work
    wanna_watch_or_watching_works.each do |work|
      channel_work = channel_works.find_by(work_id: work.id)

      create_channel_work(work) if channel_work.blank?
    end
  end

  def delete_channel_work(work)
    channel_work = channel_works.find_by(work_id: work.id)
    channel_work.destroy if channel_work.present?
  end

  def first_checkin
    checkins.order(:id).first
  end

  def first_checkin?(checkin)
    checkins.count == 1 && checkins.first.id == checkin.id
  end


  private

  def twitter_uids
    client = Twitter::REST::Client.new do |config|
      config.consumer_key        = ENV['TWITTER_CONSUMER_KEY']
      config.consumer_secret     = ENV['TWITTER_CONSUMER_SECRET']
      config.access_token        = providers.first.token
      config.access_token_secret = providers.first.token_secret
    end

    client.friend_ids.to_a.map(&:to_s)
  end

  def facebook_uids
    graph = Koala::Facebook::API.new(providers.first.token)

    graph.get_connections(:me, :friends).map { |friend| friend['id'] }
  end

  def get_large_avatar_image(provider, image_url)
    url = case provider
          when 'twitter'  then image_url.sub('_normal', '')
          when 'facebook' then "#{image_url}?type=large"
          end
    url
  end

  def publish_events
    UsersEvent.publish(:create, self)
  end
end
