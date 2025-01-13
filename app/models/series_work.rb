# typed: false
# frozen_string_literal: true

class SeriesWork < ApplicationRecord
  include DbActivityMethods
  include Unpublishable

  DIFF_FIELDS = %i[work_id].freeze

  counter_culture :series, column_name: ->(series_work) { series_work.published? ? :series_works_count : nil }

  belongs_to :series
  belongs_to :work, touch: true
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  def self.sort_season(sort_type: "ASC")
    joins(:work)
      .order("works.season_year #{sort_type}")
      .order("works.season_name #{sort_type}")
      .order("works.started_on #{sort_type}")
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) { |field, hash|
      hash[field] = send(field)
      hash
    }

    data.delete_if { |_, v| v.blank? }
  end
end
