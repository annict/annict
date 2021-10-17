# frozen_string_literal: true

# == Schema Information
#
# Table name: records
#
#  id                :bigint           not null, primary key
#  aasm_state        :string           default("published"), not null
#  deleted_at        :datetime
#  impressions_count :integer          default(0), not null
#  watched_at        :datetime         not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#  user_id           :bigint           not null
#  work_id           :bigint           not null
#
# Indexes
#
#  index_records_on_deleted_at  (deleted_at)
#  index_records_on_user_id     (user_id)
#  index_records_on_watched_at  (watched_at)
#  index_records_on_work_id     (work_id)
#
# Foreign Keys
#
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
