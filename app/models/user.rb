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

  include UserFollowable
  include UserLikable
  include UserReceivable

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
  has_many :finished_tips, dependent: :destroy
  has_many :follows,       dependent: :destroy
  has_many :followings,    through:   :follows
  has_many :r_likes,       dependent: :destroy, class_name: 'Like'
  has_many :notifications, dependent: :destroy
  has_many :providers,     dependent: :destroy
  has_many :receptions,    dependent: :destroy
  has_many :channels,      through:   :receptions
  has_many :shots,         dependent: :destroy
  has_many :statuses,      dependent: :destroy
  has_one  :profile,       dependent: :destroy

  validates :email, presence: true, uniqueness: true, email: true
  validates :username, presence: true, uniqueness: true, length: { maximum: 20 },
                       format: { with: /\A[A-Za-z0-9_]+\z/ }
  validates :terms, acceptance: true

  after_commit  :publish_events, on: :create


  def works
    @works ||= UserWorksQuery.new(self)
  end

  def episodes(work_or_works)
    UserEpisodesQuery.new(self, work_or_works)
  end

  def programs
    @programs ||= UserProgramsQuery.new(self)
  end

  def tips
    @tips ||= UserTipsQuery.new(self)
  end

  def social_friends
    @social_friends ||= UserSocialFriendsQuery.new(self)
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

  def trim_username!
    # Facebookからのユーザ登録のとき `username` に「.」が含まれている可能性があるので除去する
    username.delete!('.')
  end

  def following_activities
    following_ids = followings.pluck(:id)
    following_ids << self.id

    Activity.where(user_id: following_ids)
  end

  def watching_count
    statuses.where(latest: true).with_kind(:watching).count
  end

  def authorized_to?(provider)
    providers.pluck(:name).include?(provider.to_s)
  end

  def provider_name
    providers.first.name.humanize
  end

  def read_notifications!
    transaction do
      unread_count = notifications.unread.update_all(read: true)
      decrement!(:notifications_count, unread_count)
    end
  end

  # チャンネルと紐付いていない作品と最速放送チャンネルとを紐付ける
  def create_channel_work(work)
    channel_work = channel_works.find_by(work_id: work.id)

    if channel_work.blank?
      channel = channels.fastest(work)

      channel_works.create(work: work, channel: channel) if channel.present?
    end
  end

  # まだ番組情報が存在しなかったために `channel_works` に
  # どのチャンネルで見るかの情報が保存されなかった作品にcronで対処するためのメソッド
  def update_channel_work
    works.wanna_watch_and_watching.each do |work|
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

  def finish_tip!(tip)
    finished_tips.create!(tip: tip)
  end

  def shareable_to?(provider_name)
    providers.pluck(:name).include?(provider_name.to_s)
  end

  def twitter_client
    twitter = providers.where(name: 'twitter').first

    return nil if twitter.blank?

    Twitter::REST::Client.new do |config|
      config.consumer_key        = ENV['TWITTER_CONSUMER_KEY']
      config.consumer_secret     = ENV['TWITTER_CONSUMER_SECRET']
      config.access_token        = twitter.token
      config.access_token_secret = twitter.token_secret
    end
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
