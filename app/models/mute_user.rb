# typed: false
# frozen_string_literal: true

class MuteUser < ApplicationRecord
  belongs_to :user
  belongs_to :muted_user, class_name: "User"
end
