# typed: false
# frozen_string_literal: true

class Series < ApplicationRecord
  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[name name_en].freeze

  has_many :series_works, dependent: :destroy

  validates :name, presence: true, uniqueness: {conditions: -> { only_kept }}
end
