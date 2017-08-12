# frozen_string_literal: true
# == Schema Information
#
# Table name: users
#
#  id                        :integer          not null, primary key
#  username                  :string(510)      not null
#  email                     :string(510)      not null
#  role                      :integer          not null
#  encrypted_password        :string(510)      default(""), not null
#  remember_created_at       :datetime
#  sign_in_count             :integer          default(0), not null
#  current_sign_in_at        :datetime
#  last_sign_in_at           :datetime
#  current_sign_in_ip        :string(510)
#  last_sign_in_ip           :string(510)
#  confirmation_token        :string(510)
#  confirmed_at              :datetime
#  confirmation_sent_at      :datetime
#  unconfirmed_email         :string(510)
#  checkins_count            :integer          default(0), not null
#  notifications_count       :integer          default(0), not null
#  created_at                :datetime
#  updated_at                :datetime
#  time_zone                 :string           not null
#  locale                    :string           not null
#  reset_password_token      :string
#  reset_password_sent_at    :datetime
#  record_cache_expired_at   :datetime
#  status_cache_expired_at   :datetime
#  work_tag_cache_expired_at :datetime
#
# Indexes
#
#  users_confirmation_token_key  (confirmation_token) UNIQUE
#  users_email_key               (email) UNIQUE
#  users_username_key            (username) UNIQUE
#

class User < ApplicationRecord
  # registrations#createが実行されたあとメールアドレスの確認を挟まず
  # ログインできるようにするため、Confirmableモジュールを直接includeする
  include Devise::Models::Confirmable

  include UserCheckable
  include UserFavoritable
  include UserFollowable
  include UserLikeable
  include UserReceivable

  extend Enumerize

  attr_accessor :email_username, :current_password

  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable
  devise :database_authenticatable, :omniauthable, :registerable, :trackable,
    :rememberable, :recoverable,
    omniauth_providers: %i(facebook twitter),
    authentication_keys: %i(email_username)

  enumerize :role, in: { user: 0, admin: 1, editor: 2 }, default: :user, scope: true
  enumerize :locale, in: %i(ja en)

  has_many :activities, dependent: :destroy
  has_many :channel_works, dependent: :destroy
  has_many :records, class_name: "Checkin", dependent: :destroy
  has_many :db_activities, dependent: :destroy
  has_many :db_comments, dependent: :destroy
  has_many :favorite_characters, dependent: :destroy
  has_many :favorite_organizations, dependent: :destroy
  has_many :favorite_people, dependent: :destroy
  has_many :finished_tips, dependent: :destroy
  has_many :follows, dependent: :destroy
  has_many :followings, through: :follows
  has_many :forum_post_participants, dependent: :destroy
  has_many :latest_statuses, dependent: :destroy
  has_many :likes, dependent: :destroy
  has_many :notifications, dependent: :destroy
  has_many :providers, dependent: :destroy
  has_many :receptions, dependent: :destroy
  has_many :reviews, dependent: :destroy
  has_many :channels, through:   :receptions
  has_many :statuses, dependent: :destroy
  has_many :multiple_records, dependent: :destroy
  has_many :mute_users, dependent: :destroy
  has_many :muted_users, dependent: :destroy, foreign_key: :muted_user_id, class_name: "MuteUser"
  has_many :oauth_applications, class_name: "Doorkeeper::Application", as: :owner
  has_many :oauth_access_grants,
    class_name: "Doorkeeper::AccessGrant",
    foreign_key: :resource_owner_id,
    dependent: :destroy
  has_many :oauth_access_tokens,
    class_name: "Doorkeeper::AccessToken",
    foreign_key: :resource_owner_id,
    dependent: :destroy
  has_many :connected_applications, -> { distinct },
    class_name: "Doorkeeper::Application",
    through: :oauth_access_tokens,
    source: :application
  has_many :reactions, dependent: :destroy
  has_many :record_comments, class_name: "Comment", dependent: :destroy
  has_many :work_taggings, dependent: :destroy
  has_many :work_tags, -> { distinct }, through: :work_taggings
  has_many :work_comments, dependent: :destroy
  has_many :userland_project_members, dependent: :destroy
  has_many :userland_projects, through: :userland_project_members
  has_one :email_notification, dependent: :destroy
  has_one :profile, dependent: :destroy
  has_one :setting, dependent: :destroy

  validates :email,
    presence: true,
    uniqueness: { case_sensitive: false },
    email: true
  validates :password,
    length: { in: Devise.password_length },
    allow_blank: true,
    confirmation: { on: :password_update }
  validates :password_confirmation,
    presence: { on: :password_update }
  validates :current_password,
    valid_password: { on: :password_check }
  validates :username,
    presence: true,
    uniqueness: { case_sensitive: false },
    length: { maximum: 20 },
    format: { with: /\A[A-Za-z0-9_]+\z/ }

  # Override the Devise's `find_for_database_authentication`
  # https://github.com/plataformatec/devise/wiki/How-To:-Allow-users-to-sign-in-using-their-username-or-email-address
  def self.find_for_database_authentication(warden_conditions)
    conditions = warden_conditions.dup
    email_username = conditions.delete(:email_username)

    if email_username.present?
      where(conditions.to_h).where([
        "LOWER(email) = :value OR LOWER(username) = :value",
        value: email_username.downcase
      ]).first
    elsif conditions.key?(:email) || conditions.key?(:username)
      where(conditions.to_h).first
    end
  end

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
    unsubscription_key = "#{SecureRandom.uuid}-#{SecureRandom.uuid}"
    build_email_notification(unsubscription_key: unsubscription_key)

    self
  end

  def following_activities
    mute_user_ids = mute_users.pluck(:muted_user_id)
    following_ids = followings.where.not(id: mute_user_ids).pluck(:id)
    following_ids << self.id

    Activity.where(user_id: following_ids)
  end

  def read_notifications!
    transaction do
      unread_count = notifications.unread.update_all(read: true)
      decrement!(:notifications_count, unread_count)
    end
  end

  def authorized_to?(provider_name, shareable: false)
    records = providers
    records = records.token_available if shareable
    records.pluck(:name).include?(provider_name.to_s)
  end

  def twitter
    providers.where(name: "twitter").first
  end

  def facebook
    providers.where(name: "facebook").first
  end

  def expire_twitter_token
    return if twitter.blank?
    twitter.update_column(:token_expires_at, Time.now.to_i)
  end

  def hide_checkin_comment?(episode)
    setting.hide_checkin_comment? &&
      works.desiring_to_watch.include?(episode.work) &&
      !records.pluck(:episode_id).include?(episode.id)
  end

  def hide_review?(review)
    setting.hide_checkin_comment? &&
      works.desiring_to_watch.include?(review.work) &&
      !reviews.pluck(:work_id).include?(review.work_id)
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
    Rails.cache.fetch([id, latest_statuses, work.id]) do
      latest_statuses.find_by(work: work)&.kind.presence || "no_select"
    end
  end

  def encoded_id
    Digest::SHA256.hexdigest(id.to_s)
  end

  def mute(user)
    mute_user = mute_users.where(muted_user: user).first_or_initialize
    mute_user.save
  end

  def userland_project_member?(project)
    userland_projects.exists?(project.id)
  end

  def annict_host
    case locale
    when "ja" then ENV.fetch("ANNICT_JP_HOST")
    else
      ENV.fetch("ANNICT_HOST")
    end
  end

  def annict_url
    case locale
    when "ja" then ENV.fetch("ANNICT_JP_URL")
    else
      ENV.fetch("ANNICT_URL")
    end
  end

  def thumbs_up?(resource)
    case resource
    when CollectionItem
      reactions.where(collection_item: resource, kind: "thumbs_up").exists?
    else
      false
    end
  end

  def add_reaction!(resource, kind)
    reaction = reactions.new(kind: kind)

    case resource
    when CollectionItem
      reaction.target_user = resource.user
      reaction.collection_item = resource
    end

    reaction.save!
  end

  def remove_reaction!(resource, kind)
    reactions = self.reactions.where(kind: kind)

    case resource
    when CollectionItem
      reactions = reactions.where(collection_item: resource)
      reactions.destroy_all
    end
  end

  def add_work_tag!(work, tag_name)
    work_tag = nil

    ActiveRecord::Base.transaction do
      work_tag = WorkTag.where(name: tag_name).first_or_create!
      work_taggings.where(work: work, work_tag: work_tag).first_or_create!
      touch(:work_tag_cache_expired_at)
    end

    work_tag
  end

  def update_work_tags!(work, tag_names)
    tags = tags_by_work(work)
    removed_tag_names = tags.pluck(:name) - tag_names
    added_tag_names = tag_names - tags.pluck(:name)

    ActiveRecord::Base.transaction do
      work_tags = WorkTag.where(name: removed_tag_names)
      work_taggings.where(work: work, work_tag: work_tags).destroy_all

      added_tag_names.map do |tag_name|
        add_work_tag!(work, tag_name)
      end
    end
  end

  def tags_by_work(work)
    Rails.cache.fetch([id, work_tag_cache_expired_at, work.id]) do
      work_tags.published.joins(:work_taggings).merge(work_taggings.where(work: work))
    end
  end

  def comment_by_work(work)
    Rails.cache.fetch([id, work_comment_cache_expired_at, work.id]) do
      work_comments.find_by(work: work)
    end
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
