# frozen_string_literal: true

# == Schema Information
#
# Table name: faq_categories
#
#  id          :bigint           not null, primary key
#  aasm_state  :string           default("published"), not null
#  deleted_at  :datetime
#  locale      :string           not null
#  name        :string           not null
#  sort_number :integer          default(0), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_faq_categories_on_deleted_at  (deleted_at)
#  index_faq_categories_on_locale      (locale)
#

class FaqCategory < ApplicationRecord
  extend Enumerize

  include SoftDeletable

  enumerize :locale, in: %i[ja en]

  has_many :faq_contents, dependent: :destroy
end
