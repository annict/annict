# frozen_string_literal: true

# == Schema Information
#
# Table name: reviews
#
#  id                         :integer          not null, primary key
#  user_id                    :integer          not null
#  work_id                    :integer          not null
#  title                      :string           not null
#  body                       :text             not null
#  rating_animation_state     :string
#  rating_music_state         :string
#  rating_story_state         :string
#  rating_character_state     :string
#  rating_overall_state       :string
#  likes_count                :integer          default(0), not null
#  status_changed_users_count :integer          default(0), not null
#  impressions_count          :integer          default(0), not null
#  review_comments_count      :integer          default(0), not null
#  created_at                 :datetime         not null
#  updated_at                 :datetime         not null
#
# Indexes
#
#  index_reviews_on_user_id  (user_id)
#  index_reviews_on_work_id  (work_id)
#

class Review < ApplicationRecord
  extend Enumerize

  is_impressionable counter_cache: true, unique: true

  %i(
    rating_animation_state
    rating_music_state
    rating_story_state
    rating_character_state
    rating_overall_state
    rating_average_state
  ).each do |state|
    enumerize state, in: %i(bad average good great)
  end

  belongs_to :user
  belongs_to :work
  has_many :review_comments, dependent: :destroy

  validates :body, presence: true
  validates :rating_animation_state, presence: true
  validates :rating_character_state, presence: true
  validates :rating_music_state, presence: true
  validates :rating_overall_state, presence: true
  validates :rating_story_state, presence: true
  validates :title, presence: true, length: { maximum: 100 }
end
