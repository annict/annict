# frozen_string_literal: true

# == Schema Information
#
# Table name: collections
#
#  id           :integer          not null, primary key
#  user_id      :integer          not null
#  title        :string           not null
#  description  :string
#  aasm_state   :string           default("draft"), not null
#  likes_count  :integer          default(0), not null
#  published_at :datetime
#  created_at   :datetime         not null
#  updated_at   :datetime         not null
#
# Indexes
#
#  index_collections_on_user_id  (user_id)
#

class Collection < ApplicationRecord
  include AASM

  aasm do
    state :draft, initial: true
    state :published

    event :publish do
      transitions from: :draft, to: :published
    end
  end

  validates :title, presence: true, on: :update
  validates :description, length: { maximum: 200 }, on: :update
end
