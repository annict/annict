# frozen_string_literal: true
# == Schema Information
#
# Table name: casts
#
#  id          :integer          not null, primary key
#  person_id   :integer          not null
#  work_id     :integer          not null
#  name        :string           not null
#  part        :string           not null
#  aasm_state  :string           default("published"), not null
#  sort_number :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_casts_on_aasm_state   (aasm_state)
#  index_casts_on_person_id    (person_id)
#  index_casts_on_sort_number  (sort_number)
#  index_casts_on_work_id      (work_id)
#

class Cast < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include CastCommon

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end
end
