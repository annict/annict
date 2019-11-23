# frozen_string_literal: true
# == Schema Information
#
# Table name: faq_contents
#
#  id              :bigint           not null, primary key
#  aasm_state      :string           default("published"), not null
#  answer          :text             not null
#  deleted_at      :datetime
#  locale          :string           not null
#  question        :string           not null
#  sort_number     :integer          default(0), not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#  faq_category_id :integer          not null
#
# Indexes
#
#  index_faq_contents_on_deleted_at       (deleted_at)
#  index_faq_contents_on_faq_category_id  (faq_category_id)
#  index_faq_contents_on_locale           (locale)
#
# Foreign Keys
#
#  fk_rails_...  (faq_category_id => faq_categories.id)
#

class FaqContent < ApplicationRecord
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

  belongs_to :faq_category
end
