# typed: false
# frozen_string_literal: true

class WorkTaggable < ApplicationRecord
  belongs_to :user
  belongs_to :work_tag

  validates :description, length: {maximum: 500}
end
