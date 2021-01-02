# frozen_string_literal: true

module ListGroups
  class RecordMonthComponent < ApplicationComponent
    def initialize(user_entity:, months:)
      super
      @user_entity = user_entity
      @months = months
    end
  end
end
