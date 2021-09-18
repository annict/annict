# frozen_string_literal: true

class EpisodeRecordForm < ApplicationForm
  include FormRecordable

  attr_accessor :advanced_rating, :episode

  validates :advanced_rating, allow_nil: true, numericality: {greater_than_or_equal_to: 1, less_than_or_equal_to: 5}
  validates :episode, presence: true

  def body
    @body.presence || ""
  end
end
