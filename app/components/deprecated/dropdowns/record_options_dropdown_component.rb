# frozen_string_literal: true

module Deprecated::Dropdowns
  class RecordOptionsDropdownComponent < Deprecated::ApplicationComponent
    def initialize(viewer:, user_entity:, record_entity:)
      @viewer = viewer
      @user_entity = user_entity
      @record_entity = record_entity
    end
  end
end
