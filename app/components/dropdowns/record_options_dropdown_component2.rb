# frozen_string_literal: true

module Dropdowns
  class RecordOptionsDropdownComponent2 < ApplicationComponent
    def initialize(current_user:, record:)
      @current_user = current_user
      @record = record
    end
  end
end
