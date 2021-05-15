# frozen_string_literal: true

module V4::ListGroups
  class RecordMonthComponent < V4::ApplicationComponent
    def initialize(user_entity:, months:)
      super
      @user_entity = user_entity
      @months = months
    end
  end
end
