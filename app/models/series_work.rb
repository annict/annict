# frozen_string_literal: true
# == Schema Information
#
# Table name: series_works
#
#  id         :integer          not null, primary key
#  aasm_state :string           default("published"), not null
#  deleted_at :datetime
#  summary    :string           default(""), not null
#  summary_en :string           default(""), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  series_id  :integer          not null
#  work_id    :integer          not null
#
# Indexes
#
#  index_series_works_on_deleted_at             (deleted_at)
#  index_series_works_on_series_id              (series_id)
#  index_series_works_on_series_id_and_work_id  (series_id,work_id) UNIQUE
#  index_series_works_on_work_id                (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (series_id => series.id)
#  fk_rails_...  (work_id => works.id)
#

class SeriesWork < ApplicationRecord
  include AASM
  include DbActivityMethods

  DIFF_FIELDS = %i(work_id).freeze

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :series, counter_cache: true
  belongs_to :work
  has_many :db_activities, as: :trackable, dependent: :destroy
  has_many :db_comments, as: :resource, dependent: :destroy

  def self.sort_season(sort_type: "ASC")
    joins(:work).
      order("works.season_year #{sort_type}").
      order("works.season_name #{sort_type}")
  end

  def to_diffable_hash
    data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end

    data.delete_if { |_, v| v.blank? }
  end
end
