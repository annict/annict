# typed: false
# frozen_string_literal: true

class ForumPostParticipant < ApplicationRecord
  belongs_to :forum_post
  belongs_to :user
end
