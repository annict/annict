# typed: false
# frozen_string_literal: true

class FaqCategory < ApplicationRecord
  extend Enumerize

  include SoftDeletable

  enumerize :locale, in: %i[ja en]

  has_many :faq_contents, dependent: :destroy
end
