# typed: false
# frozen_string_literal: true

class WorkRecord < ApplicationRecord
  extend Enumerize

  include UgcLocalizable
  include SoftDeletable

  STATES = %i[
    rating_overall_state
    rating_animation_state
    rating_music_state
    rating_story_state
    rating_character_state
  ].freeze

  RATING_FIELDS = %i[
    rating_overall
    rating_animation
    rating_music
    rating_story
    rating_character
  ].freeze

  STATES.each do |state|
    enumerize state, in: Record::RATING_STATES
  end

  counter_culture :work, column_name: :work_records_count
  counter_culture :work, column_name: ->(work_record) { work_record.body.present? ? :work_records_with_body_count : nil }

  attr_accessor :mutation_error

  belongs_to :work
  belongs_to :oauth_application, class_name: "Oauth::Application", optional: true
  belongs_to :record
  belongs_to :user
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :body, length: {maximum: Record::MAX_BODY_LENGTH}

  scope :with_body, -> { where.not(body: ["", nil]) }
  scope :with_no_body, -> { where(body: ["", nil]) }
  scope :order_by_rating, ->(direction) {
    order_sql = <<~SQL
      CASE
        WHEN "work_records"."rating_overall_state" = 'bad' THEN '0'
        WHEN "work_records"."rating_overall_state" = 'average' THEN '1'
        WHEN "work_records"."rating_overall_state" = 'good' THEN '2'
        WHEN "work_records"."rating_overall_state" = 'great' THEN '3'
      END #{direction.upcase} NULLS LAST
    SQL

    order(Arel.sql(order_sql))
  }

  def comment
    body
  end

  def share_url
    "#{user.preferred_annict_url}/@#{user.username}/records/#{record.id}"
  end

  def facebook_share_title
    work.local_title
  end

  def facebook_share_body
    return body if body.present?

    if user.locale == "ja"
      "見ました。"
    else
      "Watched."
    end
  end

  def needs_single_activity_group?
    body.present?
  end
end
