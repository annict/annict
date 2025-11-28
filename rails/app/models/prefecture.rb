# typed: false

class Prefecture < ApplicationRecord
  validates :name, presence: true
end
