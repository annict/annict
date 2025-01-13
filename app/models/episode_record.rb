# typed: false
# frozen_string_literal: true

class EpisodeRecord < ApplicationRecord
  extend Enumerize

  include UgcLocalizable
  include SoftDeletable

  self.ignored_columns = %w[aasm_state multiple_episode_record_id review_id shared_facebook shared_twitter]

  enumerize :rating_state, in: Record::RATING_STATES, scope: true

  counter_culture :episode
  counter_culture :episode, column_name: ->(episode_record) { episode_record.body.present? ? :episode_record_bodies_count : nil }
  counter_culture :user

  attr_accessor :mutation_error

  belongs_to :work
  belongs_to :oauth_application, class_name: "Oauth::Application", optional: true
  belongs_to :record
  belongs_to :episode
  belongs_to :multiple_episode_record, optional: true
  belongs_to :user
  has_many :comments, dependent: :destroy
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :body, length: {maximum: 1_048_596}
  validates :rating,
    allow_blank: true,
    numericality: {
      greater_than_or_equal_to: 1,
      less_than_or_equal_to: 5
    }

  scope :with_body, -> { where.not(body: ["", nil]) }
  scope :with_no_body, -> { where(body: ["", nil]) }

  def self.initial?(record)
    count == 1 && first.id == record.id
  end

  def self.order_by_rating_state(direction = :asc)
    direction = direction.in?(%i[asc desc]) ? direction : :asc
    sql = Arel.sql(<<-SQL)
      CASE
        WHEN episode_records.rating_state = 'bad' THEN '0'
        WHEN episode_records.rating_state = 'average' THEN '1'
        WHEN episode_records.rating_state = 'good' THEN '2'
        WHEN episode_records.rating_state = 'great' THEN '3'
      END #{direction.upcase} NULLS LAST
    SQL

    order sql
  end

  def rating=(value)
    return super if value.to_f.between?(1, 5)

    write_attribute :rating, nil
  end

  def rating_to_rating_state
    case rating
    when 1.0...2.0 then :bad
    when 2.0...3.0 then :average
    when 3.0...4.0 then :good
    when 4.0..5.0 then :great
    end
  end

  def initial
    order(:id).first
  end

  def comment
    body
  end

  def generate_url_hash
    SecureRandom.urlsafe_base64.slice(0, 10)
  end

  def share_url
    "#{user.preferred_annict_url}/@#{user.username}/records/#{record.id}"
  end

  def facebook_share_title
    "#{work.title} #{episode.title_with_number}"
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
