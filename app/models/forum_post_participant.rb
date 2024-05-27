# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: forum_post_participants
#
#  id            :bigint           not null, primary key
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#  forum_post_id :bigint           not null
#  user_id       :bigint           not null
#
# Indexes
#
#  index_forum_post_participants_on_forum_post_id              (forum_post_id)
#  index_forum_post_participants_on_forum_post_id_and_user_id  (forum_post_id,user_id) UNIQUE
#  index_forum_post_participants_on_user_id                    (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (forum_post_id => forum_posts.id)
#  fk_rails_...  (user_id => users.id)
#

class ForumPostParticipant < ApplicationRecord
  belongs_to :forum_post
  belongs_to :user
end
