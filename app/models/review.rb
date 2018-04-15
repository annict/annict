# frozen_string_literal: true
# == Schema Information
#
# Table name: reviews
#
#  id                     :integer          not null, primary key
#  user_id                :integer          not null
#  work_id                :integer          not null
#  title                  :string           default("")
#  body                   :text             not null
#  rating_animation_state :string
#  rating_music_state     :string
#  rating_story_state     :string
#  rating_character_state :string
#  rating_overall_state   :string
#  likes_count            :integer          default(0), not null
#  impressions_count      :integer          default(0), not null
#  aasm_state             :string           default("published"), not null
#  modified_at            :datetime
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  oauth_application_id   :integer
#  locale                 :string           default("other"), not null
#
# Indexes
#
#  index_reviews_on_locale                (locale)
#  index_reviews_on_oauth_application_id  (oauth_application_id)
#  index_reviews_on_user_id               (user_id)
#  index_reviews_on_work_id               (work_id)
#

class Review < ApplicationRecord
  extend Enumerize
  include AASM
  include LocaleDetectable
  include Shareable

  STATES = %i(
    rating_overall_state
    rating_animation_state
    rating_music_state
    rating_story_state
    rating_character_state
  ).freeze

  is_impressionable counter_cache: true, unique: true

  STATES.each do |state|
    enumerize state, in: %i(bad average good great)
  end

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  counter_culture :work, column_name: proc { |model| model.published? ? "reviews_count" : nil }

  belongs_to :oauth_application, class_name: "Doorkeeper::Application", optional: true
  belongs_to :user
  belongs_to :work
  has_many :activities,
    dependent: :destroy,
    as: :trackable
  has_many :likes,
    dependent: :destroy,
    as: :recipient

  validates :body, length: { maximum: 1_500 }

  scope :with_body, -> { where.not(body: ["", nil]) }

  before_save :append_title_to_body

  def share_to_sns
    ShareReviewToTwitterJob.perform_later(user.id, id) if user.setting.share_review_to_twitter?
    ShareReviewToFacebookJob.perform_later(user.id, id) if user.setting.share_review_to_facebook?
  end

  # Do not use helper methods via Draper when the method is used in ActiveJob
  # https://github.com/drapergem/draper/issues/655
  def share_url
    "#{user.annict_url}/@#{user.username}/reviews/#{id}"
  end

  def facebook_share_title
    work.decorate.local_title
  end

  def twitter_share_body
    work_title = work.decorate.local_title
    title = self.body.present? ? work_title.truncate(30) : work_title
    comment = self.body.present? ? "#{self.body} / " : ""
    share_url = share_url_with_query(:twitter)
    share_hashtag = work.hashtag_with_hash

    base_body = if user.locale == "ja"
      "%s#{title} を見ました #{share_url} #{share_hashtag}"
    else
      "%sWatched: #{title} #{share_url} #{share_hashtag}"
    end

    body = base_body % comment
    body_without_url = body.sub(share_url, "")
    return body if body_without_url.length <= 130

    comment = comment.truncate(comment.length - (body_without_url.length - 130)) + " / "
    base_body % comment
  end

  def facebook_share_body
    return self.body if self.body.present?

    if user.locale == "ja"
      "見ました。"
    else
      "Watched."
    end
  end

  private

  # For backward compatible on API
  def append_title_to_body
    self.body = "#{title}\n\n#{body}" if title.present?
    self.title = ""
  end
end
