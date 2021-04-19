# frozen_string_literal: true

module Lists
  class RecordListComponent2 < ApplicationComponent
    def initialize(records:, current_user: nil, show_card: true)
      super
      @current_user = current_user
      @records = records
      @show_card = show_card
    end
  end
end
