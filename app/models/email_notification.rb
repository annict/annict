# typed: false
# frozen_string_literal: true

class EmailNotification < ApplicationRecord
  self.ignored_columns = %w[event_friends_joined]

  belongs_to :user
end
