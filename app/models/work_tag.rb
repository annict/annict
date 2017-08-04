# frozen_string_literal: true
# == Schema Information
#
# Table name: work_tags
#
#  id                   :integer          not null, primary key
#  work_tag_group_id    :integer          not null
#  name                 :string           not null
#  aasm_state           :string           default("published"), not null
#  user_work_tags_count :integer          default(0), not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#
# Indexes
#
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
end
