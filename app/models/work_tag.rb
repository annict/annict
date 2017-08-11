# frozen_string_literal: true
# == Schema Information
#
# Table name: work_tags
#
#  id                :integer          not null, primary key
#  work_tag_group_id :integer          not null
#  user_id           :integer          not null
#  name              :string           not null
#  description       :string
#  aasm_state        :string           default("published"), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#
# Indexes
#
#  index_work_tags_on_name               (name)
#  index_work_tags_on_user_id            (user_id)
#  index_work_tags_on_user_id_and_name   (user_id,name) UNIQUE
#  index_work_tags_on_work_tag_group_id  (work_tag_group_id)
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

  belongs_to :work_tag_group
  belongs_to :user
  has_many :work_taggings, dependent: :destroy

  scope :popular_tags, ->(work) {
    where(work_taggings: { work: work }).
      left_joins(:work_taggings).
      group(:id).
      select("work_tags.*, COUNT(work_taggings.id) work_taggings_count").
      order("work_taggings_count DESC")
  }
end
