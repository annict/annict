# frozen_string_literal: true
# == Schema Information
#
# Table name: staffs
#
#  id          :integer          not null, primary key
#  person_id   :integer          not null
#  work_id     :integer          not null
#  name        :string           not null
#  role        :string           not null
#  role_other  :string
#  aasm_state  :string           default("published"), not null
#  sort_number :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_staffs_on_aasm_state   (aasm_state)
#  index_staffs_on_person_id    (person_id)
#  index_staffs_on_sort_number  (sort_number)
#  index_staffs_on_work_id      (work_id)
#

class Staff < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include StaffCommon

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end
end
