# frozen_string_literal: true

# == Schema Information
#
# Table name: review_comments
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  review_id  :integer          not null
#  work_id    :integer          not null
#  body       :text             not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_review_comments_on_review_id  (review_id)
#  index_review_comments_on_user_id    (user_id)
#  index_review_comments_on_work_id    (work_id)
#

class ReviewComment < ApplicationRecord
  belongs_to :user
  belongs_to :review, counter_cache: true
  belongs_to :work

  validates :body, presence: true
end
