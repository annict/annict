# frozen_string_literal: true
# == Schema Information
#
# Table name: work_items
#
#  id         :bigint(8)        not null, primary key
#  work_id    :integer          not null
#  item_id    :integer          not null
#  user_id    :integer          not null
#  aasm_state :string           default("published"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_work_items_on_item_id              (item_id)
#  index_work_items_on_user_id              (user_id)
#  index_work_items_on_work_id              (work_id)
#  index_work_items_on_work_id_and_item_id  (work_id,item_id) UNIQUE
#

class WorkItem < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :item
  belongs_to :user
  belongs_to :work
end
