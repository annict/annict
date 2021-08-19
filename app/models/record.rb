# frozen_string_literal: true

# == Schema Information
#
# Table name: records
#
#  id                   :bigint           not null, primary key
#  aasm_state           :string           default("published"), not null
#  advanced_rating      :float
#  animation_rating     :integer
#  body                 :text             default(""), not null
#  character_rating     :integer
#  comments_count       :integer          default(0), not null
#  deleted_at           :datetime
#  facebook_url_hash    :string
#  impressions_count    :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  locale               :integer          default(0), not null
#  modified_at          :datetime
#  music_rating         :integer
#  rating               :integer
#  story_rating         :integer
#  twitter_url_hash     :string
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
  include SoftDeletable
  include Likeable

  RATING_PAIRS = {bad: 10, average: 20, good: 30, great: 40}.freeze
  RATING_KINDS = RATING_PAIRS.keys
  RATING_COLUMNS = %i[rating animation_rating character_rating music_rating story_rating].freeze

  counter_culture :user
  counter_culture :work

  RATING_COLUMNS.each do |column_name|
    enum column_name => RATING_PAIRS, _prefix: true
  end

  belongs_to :episode, optional: true
  belongs_to :user
  belongs_to :work

  scope :only_work_record, -> { where(episode_id: nil) }
  scope :only_episode_record, -> { where.not(episode_id: nil) }
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

  def local_trackable_title
    if work_record?
      return work.local_title
    end

    [work.local_title, episode_record.episode.local_number].compact.join(" ")
  end
end
