# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: series
#
#  id                 :bigint           not null, primary key
#  aasm_state         :string           default("published"), not null
#  deleted_at         :datetime
#  name               :string           not null
#  name_alter         :string           default(""), not null
#  name_alter_en      :string           default(""), not null
#  name_en            :string           default(""), not null
#  name_ro            :string           default(""), not null
#  series_works_count :integer          default(0), not null
#  unpublished_at     :datetime
#  created_at         :datetime         not null
#  updated_at         :datetime         not null
#
# Indexes
#
#  index_series_on_deleted_at      (deleted_at)
#  index_series_on_name            (name) UNIQUE
#  index_series_on_unpublished_at  (unpublished_at)
#

class Series < ApplicationRecord
  include DbActivityMethods
  include RootResourceCommon
  include Unpublishable

  DIFF_FIELDS = %i[name name_en].freeze

  has_many :series_works, dependent: :destroy

  validates :name, presence: true, uniqueness: {conditions: -> { only_kept }}
end
