# typed: false
# frozen_string_literal: true

class WorkComment < ApplicationRecord
  belongs_to :user
  belongs_to :work

  validates :body, length: {maximum: 150}, allow_blank: true
end
