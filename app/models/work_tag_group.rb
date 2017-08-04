# frozen_string_literal: true

# == Schema Information
#
# Table name: work_tag_groups
#
#  id         :integer          not null, primary key
#  name       :string           not null
#  aasm_state :string           default("published"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#

class WorkTagGroup < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end
end
