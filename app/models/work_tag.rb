# frozen_string_literal: true
# == Schema Information
#
# Table name: work_tags
#
#  id                  :bigint           not null, primary key
#  aasm_state          :string           default("published"), not null
#  locale              :string           default("other"), not null
#  name                :string           not null
#  work_taggings_count :integer          default(0), not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#
# Indexes
#
#  index_work_tags_on_locale               (locale)
#  index_work_tags_on_name                 (name) UNIQUE
#  index_work_tags_on_work_taggings_count  (work_taggings_count)
#

class WorkTag < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  has_many :work_taggables, dependent: :destroy
  has_many :work_taggings, dependent: :destroy

  scope :popular_tags, ->(work) {
    where(work_taggings: { work: work }).
      left_joins(:work_taggings).
      group(:id).
      select("work_tags.*, COUNT(work_taggings.id) work_taggings_count").
      order("work_tags.work_taggings_count DESC")
  }
end
