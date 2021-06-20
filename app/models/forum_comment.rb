# frozen_string_literal: true

# == Schema Information
#
# Table name: forum_comments
#
#  id                                                   :bigint           not null, primary key
#  body                                                 :text             not null
#  edited_at(The datetime which user has changed body.) :datetime
#  locale                                               :string           default("other"), not null
#  created_at                                           :datetime         not null
#  updated_at                                           :datetime         not null
#  forum_post_id                                        :bigint           not null
#  user_id                                              :bigint
#
# Indexes
#
#  index_forum_comments_on_forum_post_id  (forum_post_id)
#  index_forum_comments_on_locale         (locale)
#  index_forum_comments_on_user_id        (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (forum_post_id => forum_posts.id)
#  fk_rails_...  (user_id => users.id)
#

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
