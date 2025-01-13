# typed: false
# frozen_string_literal: true

class UserlandCategory < ApplicationRecord
  has_many :userland_projects
end
