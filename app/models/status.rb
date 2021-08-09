# frozen_string_literal: true

# == Schema Information
#
# Table name: statuses
#
#  id                   :bigint           not null, primary key
#  kind                 :integer          not null
#  likes_count          :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  oauth_application_id :bigint
#  user_id              :bigint           not null
#  work_id              :bigint           not null
#
# Indexes
#
#  index_statuses_on_oauth_application_id  (oauth_application_id)
#  statuses_user_id_idx                    (user_id)
#  statuses_work_id_idx                    (work_id)
#
# Foreign Keys
#
#  fk_rails_...         (oauth_application_id => oauth_applications.id)
#  statuses_user_id_fk  (user_id => users.id) ON DELETE => cascade
#  statuses_work_id_fk  (work_id => works.id) ON DELETE => cascade
#

class Status < ApplicationRecord
  extend Enumerize

  include Likeable
  include Shareable

  POSITIVE_KINDS = %i[wanna_watch watching watched].freeze

  KIND_MAPPING = {
    wanna_watch: :plan_to_watch,
    watching: :watching,
    watched: :completed,
    on_hold: :on_hold,
    stop_watching: :dropped,
    no_select: :no_status
  }.freeze

  KIND_ICONS = {
    completed: :check,
    dropped: :stop,
    on_hold: :pause,
    plan_to_watch: :circle,
    watching: :play,
    no_status: :bars
  }.freeze

  enumerize :kind, scope: true, in: {
    wanna_watch: 1,
    watching: 2,
    watched: 3,
    on_hold: 5,
    stop_watching: 4
  }

  belongs_to :activity, counter_cache: :resources_count, optional: true
  belongs_to :anime, foreign_key: :work_id
  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :user
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  scope :positive, -> { with_kind(*POSITIVE_KINDS) }
  scope :with_not_deleted_anime, -> { joins(:anime).merge(Anime.only_kept) }

  def self.kind_v2_to_v3(kind_v2)
    return if kind_v2.blank?

    KIND_MAPPING[kind_v2.to_sym]
  end

  def self.kind_v3_to_v2(kind_v3)
    return if kind_v3.blank?

    KIND_MAPPING.invert[kind_v3.to_sym]
  end

  def self.initial
    order(:id).first
  end

  def self.initial?(status)
    count == 1 && initial.id == status.id
  end

  def self.kind_icon(kind_v3)
    KIND_ICONS[kind_v3]
  end

  def self.no_status?(kind)
    kind.to_s.in?(%w[no_select no_status])
  end

  def anime_id
    work_id
  end

  def kind_v3
    self.class.kind_v2_to_v3(kind)
  end

  def share_to_sns
    return if !user.setting.share_status_to_twitter? || !user.authorized_to?(:twitter, shareable: true)

    ShareStatusToTwitterJob.perform_later(user.id, id)
  end

  def share_url
    "#{user.preferred_annict_url}/@#{user.username}/#{kind}"
  end

  def facebook_share_title
    work.local_title
  end

  def twitter_share_body
    work_title = anime.local_title
    share_url = share_url_with_query(:twitter)

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました #{share_url}"
    else
      "Changed %s's status to \"#{kind_text}\". Anime list: #{share_url}"
    end

    base_body % work_title
  end

  def facebook_share_body
    work_title = anime.local_title

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました。"
    else
      "Changed %s's status to \"#{kind_text}\"."
    end

    base_body % work_title
  end

  def needs_single_activity_group?
    false
  end

  def save_library_entry!
    library_entry = user.library_entries.find_or_initialize_by(anime: anime)
    library_entry.status = self
    library_entry.watched_episode_ids = [] if %w[watched stop_watching].include?(kind)
    library_entry.position = 1
    library_entry.save!
  end
end
