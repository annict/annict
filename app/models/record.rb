# frozen_string_literal: true

# == Schema Information
#
# Table name: records
#
#  id                   :bigint           not null, primary key
#  aasm_state           :string           default("published"), not null
#  advanced_rating      :float
#  body                 :text             default(""), not null
#  comments_count       :integer          default(0), not null
#  deleted_at           :datetime
#  impressions_count    :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  locale               :integer          default("other"), not null
#  migrated_at          :datetime
#  modified_at          :datetime
#  rating               :integer
#  recordable_type      :string
#  watched_at           :datetime
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#  episode_id           :bigint
#  oauth_application_id :bigint
#  recordable_id        :bigint
#  user_id              :bigint           not null
#  work_id              :bigint           not null
#
# Indexes
#
#  index_records_on_deleted_at                         (deleted_at)
#  index_records_on_episode_id                         (episode_id)
#  index_records_on_migrated_at                        (migrated_at)
#  index_records_on_oauth_application_id               (oauth_application_id)
#  index_records_on_recordable_id_and_recordable_type  (recordable_id,recordable_type)
#  index_records_on_user_id                            (user_id)
#  index_records_on_work_id                            (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (episode_id => episodes.id)
#  fk_rails_...  (oauth_application_id => oauth_applications.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#

class Record < ApplicationRecord
  include Likeable
  include Shareable
  include SoftDeletable
  include UgcLocalizable

  RATING_PAIRS = {bad: 10, average: 20, good: 30, great: 40}.freeze
  RATING_KINDS = RATING_PAIRS.keys
  WATCHABLE_TYPES = %w[Episode Work].freeze

  counter_culture :episode, column_name: ->(record) { record.episode_record? ? :episode_records_count : nil }
  counter_culture :episode, column_name: ->(record) { record.episode_record? && record.body.present? ? :episode_record_bodies_count : nil }
  counter_culture :user
  counter_culture :user, column_name: ->(record) { record.episode_record? ? :episode_records_count : nil }
  counter_culture :work, column_name: ->(record) { record.work_record? ? :work_records_count : nil }
  counter_culture :work, column_name: ->(record) { record.work_record? && record.body.present? ? :work_records_with_body_count : nil }

  enum rating: RATING_PAIRS, _prefix: true

  belongs_to :episode, optional: true
  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :user
  belongs_to :work
  delegated_type :recordable, types: %w[EpisodeRecord WorkRecord], dependent: :destroy
  has_many :activities, as: :itemable, dependent: :destroy

  scope :order_by_rating, ->(direction) { order("records.rating #{direction.upcase} NULLS LAST").order(advanced_rating: direction, created_at: :desc) }
  scope :with_body, -> { where.not(body: "") }
  scope :with_no_body, -> { where(body: "") }

  def liked?(likes)
    recipient_type, recipient_id = if episode_record?
      ["EpisodeRecord", episode_record.id]
    else
      ["WorkRecord", work_record.id]
    end

    likes.any? { |like| like.recipient_type == recipient_type && like.recipient_id == recipient_id }
  end

  def instant?
    rating.nil? && body.blank?
  end

  def local_trackable_title
    if work_record?
      return work.local_title
    end

    [work.local_title, episode.local_number].compact.join(" ")
  end

  def share_url
    "#{user.preferred_annict_url}/@#{user.username}/records/#{id}"
  end

  def twitter_share_body
    work_title = work.local_title
    title = body.present? ? work_title.truncate(30) : work_title
    comment = body.present? ? "#{body} / " : ""
    episode_number = episode ? "#{episode.local_number} " : ""
    share_url = share_url_with_query(:twitter)
    share_hashtag = work.hashtag_with_hash

    base_body = if user.locale == "ja"
      "%s#{title} #{episode_number}を見ました #{share_url} #{share_hashtag}"
    else
      "%sWatched: #{title} #{episode_number}#{share_url} #{share_hashtag}"
    end

    body = base_body % comment
    body_without_url = body.sub(share_url, "")
    return body if body_without_url.length <= 130

    comment = comment.truncate(comment.length - (body_without_url.length - 130)) + " / "
    base_body % comment
  end

  def needs_single_activity_group?
    body.present?
  end
end
