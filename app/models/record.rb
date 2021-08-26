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
#  locale               :integer          default(0), not null
#  migrated_at          :datetime
#  modified_at          :datetime
#  rating               :integer
#  watched_at           :datetime
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#  episode_id           :bigint
#  oauth_application_id :bigint
#  user_id              :bigint           not null
#  work_id              :bigint           not null
#
# Indexes
#
#  index_records_on_deleted_at            (deleted_at)
#  index_records_on_episode_id            (episode_id)
#  index_records_on_migrated_at           (migrated_at)
#  index_records_on_oauth_application_id  (oauth_application_id)
#  index_records_on_user_id               (user_id)
#  index_records_on_work_id               (work_id)
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
  RATING_COLUMNS = %i[rating animation_rating character_rating music_rating story_rating].freeze
  WATCHABLE_TYPES = %w[Episode Work].freeze

  self.ignored_columns = %w[
    work_id
  ]

  counter_culture :user

  RATING_COLUMNS.each do |column_name|
    enum column_name => RATING_PAIRS, _prefix: true
  end

  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :user

  scope :on_work, -> { where(episode_id: nil) }
  scope :on_episode, -> { where.not(episode_id: nil) }
  scope :order_by_rating, ->(direction) { order("records.rating #{direction.upcase} NULLS LAST").order(advanced_rating: direction, created_at: :desc) }
  scope :with_body, -> { where.not(body: "") }
  scope :with_no_body, -> { where(body: "") }

  def work_record?
    episode_id.nil?
  end

  def episode_record?
    !work_record?
  end

  def deprecated_rating_exists?
    animation_rating || music_rating || story_rating || character_rating
  end

  def comment
    episode_record? ? episode_record.body : work_record&.body
  end

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
