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
  include SoftDeletable
  include Likeable

  RATING_STATES = %i[bad average good great].freeze

  counter_culture :user
  counter_culture :work

  belongs_to :work
  belongs_to :user
  has_one :work_record, dependent: :destroy
  has_one :episode_record, dependent: :destroy

  scope :with_work_record, -> { joins(:work_record).merge(WorkRecord.only_kept) }

  def episode_record?
    episode_record.present?
  end

  def work_record?
    !episode_record?
  end

  def episode
    episode_record? ? episode_record.episode : nil
  end

  def rating
    episode_record? ? episode_record.rating_state : work_record&.rating_overall_state
  end

  def advanced_rating
    episode_record? ? episode_record.rating : nil
  end

  def deprecated_animation_rating
    work_record&.rating_animation_state
  end

  def deprecated_character_rating
    work_record&.rating_character_state
  end

  def deprecated_music_rating
    work_record&.rating_music_state
  end

  def deprecated_story_rating
    work_record&.rating_story_state
  end

  def deprecated_rating_exists?
    deprecated_animation_rating || deprecated_music_rating || deprecated_story_rating || deprecated_character_rating
  end

  def comment
    episode_record? ? episode_record.body : work_record&.body
  end

  def modified_at
    if work_record?
      return work_record&.modified_at
    end

    episode_record.updated_at if episode_record.modify_body?
  end

  def likes_count
    episode_record? ? episode_record.likes_count : work_record&.likes_count
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
