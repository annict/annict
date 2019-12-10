# frozen_string_literal: true
# == Schema Information
#
# Table name: users
#
#  id                            :integer          not null, primary key
#  aasm_state                    :string           default("published"), not null
#  allowed_locales               :string           is an Array
#  confirmation_sent_at          :datetime
#  confirmation_token            :string(510)
#  confirmed_at                  :datetime
#  current_sign_in_at            :datetime
#  current_sign_in_ip            :string(510)
#  deleted_at                    :datetime
#  email                         :citext           not null
#  encrypted_password            :string(510)      default(""), not null
#  episode_records_count         :integer          default(0), not null
#  last_sign_in_at               :datetime
#  last_sign_in_ip               :string(510)
#  locale                        :string           not null
#  notifications_count           :integer          default(0), not null
#  record_cache_expired_at       :datetime
#  records_count                 :integer          default(0), not null
#  remember_created_at           :datetime
#  reset_password_sent_at        :datetime
#  reset_password_token          :string
#  role                          :integer          not null
#  sign_in_count                 :integer          default(0), not null
#  status_cache_expired_at       :datetime
#  time_zone                     :string           not null
#  unconfirmed_email             :string(510)
#  username                      :citext           not null
#  work_comment_cache_expired_at :datetime
#  work_tag_cache_expired_at     :datetime
#  created_at                    :datetime
#  updated_at                    :datetime
#  gumroad_subscriber_id         :integer
#
# Indexes
#
#  index_users_on_aasm_state             (aasm_state)
#  index_users_on_allowed_locales        (allowed_locales) USING gin
#  index_users_on_deleted_at             (deleted_at)
#  index_users_on_gumroad_subscriber_id  (gumroad_subscriber_id)
#  users_confirmation_token_key          (confirmation_token) UNIQUE
#  users_email_key                       (email) UNIQUE
#  users_username_key                    (username) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (gumroad_subscriber_id => gumroad_subscribers.id)
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
  include SoftDeletable

  extend Enumerize

  attr_accessor :email_username, :current_password, :terms_and_privacy_policy_agreement

  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable
  devise :database_authenticatable, :omniauthable, :registerable, :trackable,
    :rememberable, :recoverable,
    omniauth_providers: %i(facebook gumroad twitter),
    authentication_keys: %i(email_username)

  enumerize :allowed_locales, in: ApplicationRecord::LOCALES, multiple: true, default: ApplicationRecord::LOCALES
  enumerize :locale, in: %i(ja en)
  enumerize :role, in: { user: 0, admin: 1, editor: 2 }, default: :user, scope: true

  belongs_to :gumroad_subscriber, optional: true
  has_many :activities, dependent: :destroy
  has_many :channel_works, dependent: :destroy
  has_many :episode_records, dependent: :destroy
  has_many :db_activities, dependent: :destroy
  has_many :db_comments, dependent: :destroy
  has_many :favorite_characters, dependent: :destroy
  has_many :favorite_organizations, dependent: :destroy
  has_many :favorite_people, dependent: :destroy
  has_many :finished_tips, dependent: :destroy
  has_many :follows, dependent: :destroy
  has_many :followings, through: :follows
  has_many :forum_post_participants, dependent: :destroy
  has_many :library_entries, dependent: :destroy
  has_many :likes, dependent: :destroy
  has_many :notifications, dependent: :destroy
  has_many :providers, dependent: :destroy
  has_many :receptions, dependent: :destroy
  has_many :channels, through:   :receptions
  has_many :statuses, dependent: :destroy
  has_many :multiple_episode_records, dependent: :destroy
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
  has_many :records, dependent: :destroy
  has_many :work_taggables, dependent: :destroy
  has_many :work_taggings, dependent: :destroy
  has_many :work_tags, through: :work_taggables
  has_many :work_comments, dependent: :destroy
  has_many :work_records, dependent: :destroy
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
    length: { maximum: 20 },
    format: { with: /\A[A-Za-z0-9_]+\z/ },
    uniqueness: { case_sensitive: false }
  validates :terms_and_privacy_policy_agreement, acceptance: true

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

  def works
    @works ||= UserWorksQuery.new(self)
  end

  def works_on(*status_kinds)
    Work.joins(:library_entries).merge(library_entries.with_status(*status_kinds))
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
        p.name = oauth["provider"]
        p.uid = oauth["uid"]
        p.token = oauth["credentials"]["token"]
        p.token_expires_at = oauth["credentials"]["expires_at"]
        p.token_secret = oauth["credentials"]["secret"]
      end

      build_profile do |p|
        p.name = oauth["info"]["name"].presence || oauth["info"]["nickname"]
        p.description = oauth["info"]["description"]
        image_url = get_large_avatar_image(oauth["provider"], oauth["info"]["image"])
        p.image = Down.open(image_url)
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

  def gumroad
    providers.where(name: "gumroad").first
  end

  def expire_twitter_token
    return if twitter.blank?
    twitter.update_column(:token_expires_at, Time.now.to_i)
  end

  def hide_episode_record_body?(episode)
    setting.hide_record_body? &&
      works.desiring_to_watch.include?(episode.work) &&
      !episode_records.pluck(:episode_id).include?(episode.id)
  end

  def hide_work_record_body?(work)
    setting.hide_record_body? &&
      works.desiring_to_watch.include?(work) &&
      !work_records.pluck(:work_id).include?(work.id)
  end

  def committer?
    role.admin? || role.editor?
  end

  def friends_interested_in(work)
    status_kinds = %w(wanna_watch watching watched)
    library_entries = LibraryEntry.where(work: work).with_status(*status_kinds)

    followings.joins(:library_entries).merge(library_entries)
  end

  def status_kind(work)
    library_entries.find_by(work: work)&.status&.kind.presence || "no_select"
  end

  def status_kind_v3(work)
    Status.kind_v2_to_v3(library_entries.find_by(work: work)&.status&.kind)&.to_s.presence || "no_status"
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

  def preferred_annict_url
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
      work_taggables.where(work_tag: work_tag).first_or_create!
      work_taggings.where(work: work, work_tag: work_tag).first_or_create!
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

      work_tags.each do |work_tag|
        unless work_taggings.where(work_tag: work_tag).exists?
          taggable = work_taggables.find_by(work_tag: work_tag)
          taggable.destroy if taggable.present?
        end
      end

      added_tag_names.map do |tag_name|
        add_work_tag!(work, tag_name)
      end
    end
  end

  def tags_by_work(work)
    work_tags.without_deleted.joins(:work_taggings).merge(work_taggings.where(work: work))
  end

  def comment_by_work(work)
    work_comments.find_by(work: work)
  end

  def supporter?
    gumroad_subscriber.present? &&
      (gumroad_subscriber.gumroad_ended_at.nil? || gumroad_subscriber.gumroad_ended_at > Time.zone.now)
  end

  def weeks
    days = (Time.zone.now.to_date - created_at.to_date).to_f
    (days / 7).floor
  end

  def leave
    username = SecureRandom.uuid.tr("-", "_")

    ActiveRecord::Base.transaction do
      update_columns(username: username, email: "#{username}@example.com", deleted_at: Time.zone.now)
      providers.delete_all

      oauth_applications.available.find_each do |app|
        app.update(owner: nil)
        app.soft_delete
      end
    end
  end

  def slot_data(library_entries)
    channel_works = self.channel_works.where(work_id: library_entries.pluck(:work_id))
    channel_ids = channel_works.pluck(:channel_id)
    episode_ids = library_entries.pluck(:next_episode_id)
    slots = Slot.
      includes(:channel, work: :work_image).
      where(channel_id: channel_ids, episode_id: episode_ids).
      without_deleted

    channel_works.map do |cw|
      slot = slots.
        select { |p| p.work_id == cw.work_id && p.channel_id == cw.channel_id }.
        sort_by(&:started_at).
        reverse.
        first

      slot
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
