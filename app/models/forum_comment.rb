# frozen_string_literal: true
# == Schema Information
#
# Table name: forum_comments
#
#  id            :integer          not null, primary key
#  user_id       :integer          not null
#  forum_post_id :integer          not null
#  body          :text             not null
#  edited_at     :datetime
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#
# Indexes
#
#  index_forum_comments_on_forum_post_id  (forum_post_id)
#  index_forum_comments_on_user_id        (user_id)
#

class ForumComment < ApplicationRecord
  belongs_to :forum_post, counter_cache: true
  belongs_to :user

  validates :body, presence: true, length: { maximum: 5000 }
  validates :forum_post, presence: true
  validates :user, presence: true
end
