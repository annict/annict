# typed: false
# frozen_string_literal: true

class Reaction < ApplicationRecord
  extend Enumerize

  enumerize :kind, in: %w[
    thumbs_up
  ]

  validates :kind, presence: true

  belongs_to :user
  belongs_to :target_user, class_name: "User"
end
