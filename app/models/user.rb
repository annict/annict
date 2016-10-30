# frozen_string_literal: true
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
#  time_zone            :string           default(""), not null
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

  include UserCheckable
  include UserFollowable
  include UserLikable
  include UserReceivable

  extend Enumerize

  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :recoverable, :rememberable
  devise :database_authenticatable, :omniauthable, :registerable,
         :trackable, omniauth_providers: [:facebook, :twitter]

  enumerize :role, in: { user: 0, admin: 1, editor: 2 }, default: :user, scope: true

  has_many :activities,    dependent: :destroy
  has_many :channel_works, dependent: :destroy
  has_many :checkins,      dependent: :destroy
  has_many :db_comments, dependent: :destroy
  has_many :finished_tips, dependent: :destroy
  has_many :follows,       dependent: :destroy
  has_many :followings,    through:   :follows
  has_many :latest_statuses, dependent: :destroy
  has_many :r_likes,       dependent: :destroy, class_name: 'Like'
  has_many :notifications, dependent: :destroy
  has_many :providers,     dependent: :destroy
  has_many :receptions,    dependent: :destroy
  has_many :channels,      through:   :receptions
  has_many :statuses,      dependent: :destroy
  has_many :multiple_records, dependent: :destroy
  has_many :mute_users, dependent: :destroy
  has_many :oauth_applications, class_name: "Doorkeeper::Application", as: :owner
  has_many :oauth_access_grants,
    class_name: "Doorkeeper::AccessGrant",
    foreign_key: :resource_owner_id,
    dependent: :destroy
  has_many :oauth_access_tokens,
    class_name: "Doorkeeper::AccessToken",
    foreign_key: :resource_owner_id,
    dependent: :destroy
  has_one  :profile,       dependent: :destroy
  has_one  :setting,       dependent: :destroy

  validates :email, presence: true, uniqueness: true, email: true
  validates :username, presence: true, uniqueness: true, length: { maximum: 20 },
                       format: { with: /\A[A-Za-z0-9_]+\z/ }

  def checking_works
    @checking_works ||= CheckingWorks.new(read_attribute(:checking_works))
  end

  def works
    @works ||= UserWorksQuery.new(self)
  end

  def episodes
    @episodes ||= UserEpisodesQuery.new(self)
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

  def build_relations(oauth = nil)
    if oauth.present?
      providers.build do |p|
        p.name = oauth[:provider]
        p.uid = oauth[:uid]
        p.token = oauth[:credentials][:token]
        p.token_expires_at = oauth[:credentials][:expires_at]
        p.token_secret = oauth[:credentials][:secret]
      end

      build_profile do |p|
        p.name = oauth[:info][:name].presence || oauth[:info][:nickname]
        p.description = oauth[:info][:description]
        image_url = get_large_avatar_image(oauth[:provider], oauth[:info][:image])
        p.tombo_avatar = URI.parse(image_url)
      end
    else
      build_profile(name: username)
    end

    build_setting

    self
  end

  def following_activities
    mute_user_ids = mute_users.pluck(:muted_user_id)
    following_ids = followings.where.not(id: mute_user_ids).pluck(:id)
    following_ids << self.id

    Activity.where(user_id: following_ids)
  end

  def authorized_to?(provider)
    providers.pluck(:name).include?(provider.to_s)
  end

  def read_notifications!
    transaction do
      unread_count = notifications.unread.update_all(read: true)
      decrement!(:notifications_count, unread_count)
    end
  end

  def shareable_to?(provider_name)
    providers.pluck(:name).include?(provider_name.to_s)
  end

  def twitter
    providers.where(name: "twitter").first
  end

  def facebook
    providers.where(name: "facebook").first
  end

  def hide_checkin_comment?(checkin)
    checkin.comment.present? &&
    checkin.user != self &&
    setting.hide_checkin_comment? &&
    works.desiring_to_watch.include?(checkin.episode.work) &&
    !checkins.pluck(:episode_id).include?(checkin.episode_id)
  end

  def committer?
    role.admin? || role.editor?
  end

  def friends_interested_in(work)
    status_kinds = %w(wanna_watch watching watched)
    latest_statuses = LatestStatus.where(work: work).with_kind(*status_kinds)

    followings.joins(:latest_statuses).merge(latest_statuses)
  end

  def status_kind(work)
    latest_statuses.find_by(work: work)&.kind.presence || "no_select"
  end

  def ga_uid
    Digest::SHA256.hexdigest(id.to_s)
  end

  def mute(user)
    mute_user = mute_users.where(muted_user: user).first_or_initialize
    mute_user.save
  end

  private

  def get_large_avatar_image(provider, image_url)
    url = case provider
          when 'twitter'  then image_url.sub('_normal', '')
          when 'facebook' then "#{image_url.sub("http://", "https://")}?type=large"
          end
    url
  end
end
