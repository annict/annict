# frozen_string_literal: true

class RecordFooterComponent < ApplicationComponent
  def initialize(record:)
    @record = record
  end
end
