# typed: false
# frozen_string_literal: true

class ForumComment < ApplicationRecord
  include UgcLocalizable

  counter_culture :forum_post

  belongs_to :forum_post
  belongs_to :user, optional: true

  validates :body, presence: true, length: {maximum: 5000}
  validates :forum_post, presence: true
  validates :user, presence: true

  def send_notification
    forum_post.forum_post_participants.where.not(user: user).each do |p|
      ForumMailer.comment_notification(p.user.id, id).deliver_later
    end
  end
end
