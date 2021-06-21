# frozen_string_literal: true

# == Schema Information
#
# Table name: records
#
#  id                :bigint           not null, primary key
#  aasm_state        :string           default("published"), not null
#  deleted_at        :datetime
#  impressions_count :integer          default(0), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#  user_id           :bigint           not null
#  work_id           :bigint           not null
#
# Indexes
#
#  index_records_on_deleted_at  (deleted_at)
#  index_records_on_user_id     (user_id)
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

  attr_accessor :is_liked, :is_spoiler

  counter_culture :user
  counter_culture :work

  belongs_to :anime, foreign_key: :work_id
  belongs_to :user
  belongs_to :work
  has_one :anime_record, dependent: :destroy
  has_one :episode_record, dependent: :destroy

  scope :with_anime_record, -> { joins(:anime_record).merge(AnimeRecord.only_kept) }
  scope :order_by_rating, ->(direction) {
    order_sql = <<~SQL
      CASE
        WHEN "work_records"."rating_overall_state" = 'bad' THEN '0'
        WHEN "work_records"."rating_overall_state" = 'average' THEN '1'
        WHEN "work_records"."rating_overall_state" = 'good' THEN '2'
        WHEN "work_records"."rating_overall_state" = 'great' THEN '3'
      END #{direction.upcase} NULLS LAST
    SQL

    joins(:anime_record).order(Arel.sql(order_sql)).order("records.created_at": :desc)
  }

  def anime_id
    work_id
  end

  def episode_record?
    episode_record.present?
  end

  def anime_record?
    !episode_record?
  end

  def modified_at
    if anime_record?
      return anime_record.modified_at
    end

    episode_record.updated_at if episode_record.modify_body?
  end

  def likes_count
    episode_record? ? episode_record.likes_count : anime_record.likes_count
  end

  def liked?(likes)
    recipient_type, recipient_id = if episode_record?
      ["EpisodeRecord", episode_record.id]
    else
      ["WorkRecord", anime_record.id]
    end

    likes.any? { |like| like.recipient_type == recipient_type && like.recipient_id == recipient_id }
  end
end
