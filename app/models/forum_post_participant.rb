# frozen_string_literal: true
# == Schema Information
#
# Table name: forum_post_participants
#
#  id            :integer          not null, primary key
#  forum_post_id :integer          not null
#  user_id       :integer          not null
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#
# Indexes
#
#  index_forum_post_participants_on_forum_post_id              (forum_post_id)
#  index_forum_post_participants_on_forum_post_id_and_user_id  (forum_post_id,user_id) UNIQUE
#  index_forum_post_participants_on_user_id                    (user_id)
#

class ForumPostParticipant < ApplicationRecord
  belongs_to :forum_post
  belongs_to :user
end
