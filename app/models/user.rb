# frozen_string_literal: true

# == Schema Information
#
# Table name: users
#
#  id                            :bigint           not null, primary key
#  aasm_state                    :string           default("published"), not null
#  allowed_locales               :string           is an Array
#  character_favorites_count     :integer          default(0), not null
#  completed_works_count         :integer          default(0), not null
#  confirmation_sent_at          :datetime
#  confirmation_token            :string(510)
#  confirmed_at                  :datetime
#  current_sign_in_at            :datetime
#  current_sign_in_ip            :string(510)
#  deleted_at                    :datetime
#  dropped_works_count           :integer          default(0), not null
#  email                         :citext           not null
#  encrypted_password            :string(510)      default(""), not null
#  episode_records_count         :integer          default(0), not null
#  followers_count               :integer          default(0), not null
#  following_count               :integer          default(0), not null
#  last_sign_in_at               :datetime
#  last_sign_in_ip               :string(510)
#  locale                        :string           not null
#  notifications_count           :integer          default(0), not null
#  on_hold_works_count           :integer          default(0), not null
#  organization_favorites_count  :integer          default(0), not null
#  person_favorites_count        :integer          default(0), not null
#  plan_to_watch_works_count     :integer          default(0), not null
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
#  watching_works_count          :integer          default(0), not null
#  work_comment_cache_expired_at :datetime
#  work_tag_cache_expired_at     :datetime
#  created_at                    :datetime
#  updated_at                    :datetime
#  gumroad_subscriber_id         :bigint
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
  include UserCheckable
  include UserFavoritable
  include UserFollowable
  include UserLikeable
  include UserReceivable
  include SoftDeletable

  extend Enumerize

  USERNAME_FORMAT = /\A[A-Za-z0-9_]+\z/

  attr_accessor :email_username, :current_password

  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable
  devise :database_authenticatable, :omniauthable, :registerable, :trackable,
    :rememberable, :recoverable,
    omniauth_providers: %i[facebook gumroad twitter],
    authentication_keys: %i[email_username]

  enumerize :allowed_locales, in: ApplicationRecord::LOCALES, multiple: true, default: ApplicationRecord::LOCALES
  enumerize :locale, in: %i[ja en]
  enumerize :role, in: {user: 0, admin: 1, editor: 2}, default: :user, scope: true

  belongs_to :gumroad_subscriber, optional: true
  has_many :activity_groups, dependent: :destroy
  has_many :activities, dependent: :destroy
  has_many :work_records, dependent: :destroy
  has_many :character_favorites, dependent: :destroy
  has_many :collections, dependent: :destroy
  has_many :collection_items, dependent: :destroy
  has_many :email_confirmations, dependent: :destroy
  has_many :organization_favorites, dependent: :destroy
  has_many :person_favorites, dependent: :destroy
  has_many :favorite_characters, through: :character_favorites, source: :character
  has_many :favorite_organizations, through: :organization_favorites, source: :organization
  has_many :favorite_people, through: :person_favorites, source: :person
  has_many :episode_records, dependent: :destroy
  has_many :db_activities, dependent: :destroy
  has_many :db_comments, dependent: :destroy
  has_many :follows, dependent: :destroy
  has_many :followings, through: :follows
  has_many :forum_comments, dependent: :nullify
  has_many :forum_post_participants, dependent: :destroy
  has_many :forum_posts, dependent: :destroy
  has_many :library_entries, dependent: :destroy
  has_many :likes, dependent: :destroy
  has_many :notifications, dependent: :destroy
  has_many :providers, dependent: :destroy
  has_many :receptions, dependent: :destroy
  has_many :channels, through: :receptions
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

  delegate :name, to: :profile
  delegate :admin?, :editor?, to: :role
  delegate :hide_record_body?, :hide_supporter_badge?, :share_record_to_twitter?, to: :setting

  validates :email,
    presence: true,
    uniqueness: {case_sensitive: false},
    email: true
  validates :password,
    length: {in: Devise.password_length},
    allow_blank: true,
    confirmation: {on: :password_update}
  validates :password_confirmation,
    presence: {on: :password_update}
  validates :current_password,
    valid_password: {on: :password_check}
  validates :username,
    presence: true,
    length: {maximum: 20},
    format: {with: USERNAME_FORMAT},
    uniqueness: {case_sensitive: false}

  # Override the Devise's `find_for_database_authentication`
  # https://github.com/plataformatec/devise/wiki/How-To:-Allow-users-to-sign-in-using-their-username-or-email-address
  def self.find_for_database_authentication(warden_conditions)
    conditions = warden_conditions.dup
    email_username = conditions.delete(:email_username)

    if email_username.present?
      where(conditions.to_h).where([
        "LOWER(email) = :value OR LOWER(username) = :value",
        {value: email_username.downcase}
      ]).first
    elsif conditions.key?(:email) || conditions.key?(:username)
      where(conditions.to_h).first
    end
  end

  def watching_work_count
    watching_works_count
  end

  def works
    Work.joins(:library_entries).merge(library_entries)
  end

  def works_on(*status_kinds)
    Work.joins(:library_entries).merge(library_entries.with_status(*status_kinds))
  end

  def cast_favorites
    person_favorites.with_cast
  end

  def staff_favorites
    person_favorites.with_staff
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

  def read_notifications!
    transaction do
      unread_count = notifications.unread.update_all(read: true)
      decrement!(:notifications_count, unread_count)
    end
  end

  def save_program_to_library_entry!(work, program)
    library_entry = library_entries.find_or_initialize_by(work: work)

    if program
      library_entry.program = program
      library_entry.set_next_resources!
    else
      library_entry.program = nil
      library_entry.next_slot = nil
      library_entry.position = 1 if library_entry.status&.kind&.watching?
    end

    library_entry.save!
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
    admin? || editor?
  end

  def friends_interested_in(work)
    status_kinds = %w[wanna_watch watching watched]
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
    when "ja" then ENV.fetch("ANNICT_HOST")
    else
      ENV.fetch("ANNICT_EN_HOST")
    end
  end

  def preferred_annict_url
    case locale
    when "ja" then ENV.fetch("ANNICT_URL")
    else
      ENV.fetch("ANNICT_EN_URL")
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

  def add_work_tag!(work, tag_name)
    work_tag = nil

    ActiveRecord::Base.transaction do
      work_tag = WorkTag.where(name: tag_name).first_or_create!
      work_taggables.where(work_tag: work_tag).first_or_create!
      work_taggings.where(work: work, work_tag: work_tag).first_or_create!
    end

    work_tag
  end

  def comment_by_work(work)
    work_comments.find_by(work: work)
  end

  def supporter?
    gumroad_subscriber&.active? == true
  end

  def weeks
    days = (Time.zone.now.to_date - created_at.to_date).to_f
    (days / 7).floor
  end

  def days_from_started(time_zone)
    ((Time.zone.now.in_time_zone(time_zone) - created_at.in_time_zone(time_zone)) / 86_400).ceil
  end

  def validate_to_destroy
    if oauth_applications.where(deleted_at: nil).exists?
      errors.add(:oauth_applications, I18n.t("messages.users._validators.exists_active_oauth_applications"))
      return false
    end

    true
  end

  def registered_after_email_confirmation_required?
    # After 2020-04-07, registered users must confirm email before sign in
    created_at > Time.zone.parse("2020-04-07 0:00:00")
  end

  def create_or_last_activity_group!(itemable)
    itemable_type = itemable.class.name

    if itemable.needs_single_activity_group?
      return activity_groups.create!(itemable_type: itemable_type, single: true)
    end

    last_activity_group = activity_groups.after(Time.zone.now - 12.hours).order(created_at: :desc).first

    if last_activity_group&.itemable_type == itemable_type && !last_activity_group.single?
      return last_activity_group
    end

    activity_groups.create!(itemable_type: itemable_type, single: false)
  end

  def update_works_count!(prev_status_kind, next_status_kind)
    works_count_fields = {
      wanna_watch: :plan_to_watch_works_count,
      watching: :watching_works_count,
      watched: :completed_works_count,
      on_hold: :on_hold_works_count,
      stop_watching: :dropped_works_count
    }.freeze
    prev_no_status = Status.no_status?(prev_status_kind)
    next_no_status = Status.no_status?(next_status_kind)

    decrement!(works_count_fields[prev_status_kind]) unless prev_no_status
    increment!(works_count_fields[next_status_kind]) unless next_no_status
  end

  def update_share_record_setting(share_to_twitter)
    return if share_to_twitter == setting.share_record_to_twitter

    setting.update_column(:share_record_to_twitter, share_to_twitter)
  end

  def share_episode_record_to_twitter(episode_record)
    return unless share_record_to_twitter?

    ShareEpisodeRecordToTwitterJob.perform_later(id, episode_record.id)
  end

  def share_work_record_to_twitter(work_record)
    return unless share_record_to_twitter?

    ShareWorkRecordToTwitterJob.perform_later(id, work_record.id)
  end

  def following_resources(model: Activity, viewer: nil, order: OrderProperty.new)
    target_user_ids = followings.only_kept.pluck(:id)
    target_user_ids -= viewer&.mute_users&.pluck(:muted_user_id).presence || []
    target_user_ids << id
    target_users = User.where(id: target_user_ids).only_kept

    resources = model.joins(:user).merge(target_users)

    resources.order(order.field => order.direction)
  end

  def following_user_ids
    user_ids = followings.only_kept.pluck(:id)
    user_ids -= mute_users&.pluck(:muted_user_id).presence || []
    user_ids << id
    user_ids
  end

  def confirm_to_update_email!(new_email:)
    email_confirmations.new(email: new_email).confirm_to_update_email!
  end

  def confirm
    touch :confirmed_at
  end

  def confirmed?
    !!confirmed_at
  end

  def last_record_watched_at
    records.select(:created_at).last&.created_at
  end

  def filter_records(base_record_entities, record_entities)
    muted_user_ids = mute_users.pluck(:muted_user_id)
    record_ids = record_entities.pluck(:database_id)

    base_record_entities.filter do |record_entity|
      user_id = record_entity.user.database_id
      record_id = record_entity.database_id

      !user_id.in?(muted_user_ids) && !record_id.in?(record_ids)
    end
  end

  private

  def get_large_avatar_image(provider, image_url)
    case provider
    when "twitter" then image_url.sub("_normal", "")
    when "facebook" then "#{image_url.sub("http://", "https://")}?type=large"
    end
  end
end
