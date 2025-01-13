# typed: false
# frozen_string_literal: true

class FaqContent < ApplicationRecord
  extend Enumerize

  include SoftDeletable

  enumerize :locale, in: %i[ja en]

  belongs_to :faq_category
end
