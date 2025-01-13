# typed: false
# frozen_string_literal: true

class WorkTag < ApplicationRecord
  include SoftDeletable

  has_many :work_taggables, dependent: :destroy
  has_many :work_taggings, dependent: :destroy

  scope :popular_tags, ->(work) {
    where(work_taggings: {work: work})
      .left_joins(:work_taggings)
      .group(:id)
      .select("work_tags.*, COUNT(work_taggings.id) work_taggings_count")
      .order("work_tags.work_taggings_count DESC")
  }
end
