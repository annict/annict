# frozen_string_literal: true

module Deprecated::ListGroups
  class RecordMonthComponent < Deprecated::ApplicationComponent
    def initialize(user_entity:, months:)
      super
      @user_entity = user_entity
      @months = months
    end
  end
end
