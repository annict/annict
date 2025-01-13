# typed: false
# frozen_string_literal: true

class UserlandProjectMember < ApplicationRecord
  belongs_to :user
  belongs_to :userland_project
end
