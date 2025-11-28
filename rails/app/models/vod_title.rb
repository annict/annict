# typed: false
# frozen_string_literal: true

class VodTitle < ApplicationRecord
  include SoftDeletable

  belongs_to :channel
  belongs_to :work, optional: true

  def import_csv
    [
      channel_id,
      nil,
      nil,
      code,
      name
    ].join(",")
  end
end
