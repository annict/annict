# frozen_string_literal: true
# == Schema Information
#
# Table name: collections
#
#  id                :integer          not null, primary key
#  user_id           :integer          not null
#  name              :string           not null
#  description       :string
#  aasm_state        :string           default("published"), not null
#  likes_count       :integer          default(0), not null
#  impressions_count :integer          default(0), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#
# Indexes
#
#  index_collections_on_user_id  (user_id)
#

class Collection < ApplicationRecord
  include AASM

  is_impressionable counter_cache: true, unique: true

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :user
  has_many :collection_items, dependent: :destroy

  validates :title, presence: true, length: { maximum: 50 }
  validates :description, length: { maximum: 500 }
end
