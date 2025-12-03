# typed: false
# frozen_string_literal: true

class Series < ApplicationRecord
  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[name name_en].freeze

  has_many :series_works, dependent: :destroy

  validates :name, presence: true, uniqueness: {conditions: -> { only_kept }}

  def self.ransackable_attributes(auth_object = nil)
    %w[name name_ro name_en]
  end

  def self.ransackable_associations(auth_object = nil)
    []
  end
end
