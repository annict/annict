# frozen_string_literal: true
# == Schema Information
#
# Table name: faq_categories
#
#  id          :bigint           not null, primary key
#  aasm_state  :string           default("published"), not null
#  locale      :string           not null
#  name        :string           not null
#  sort_number :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_faq_categories_on_locale  (locale)
#

class FaqCategory < ApplicationRecord
  extend Enumerize
  include AASM

  enumerize :locale, in: %i(ja en)

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  has_many :faq_contents, dependent: :destroy
end
