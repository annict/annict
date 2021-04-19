# frozen_string_literal: true

class RecordHeaderComponent2 < ApplicationComponent
  def initialize(record:, current_user: nil)
    @current_user = current_user
    @record = record
  end
end
