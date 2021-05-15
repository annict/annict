# frozen_string_literal: true

module V4::Dropdowns
  class RecordOptionsDropdownComponent < V4::ApplicationComponent
    def initialize(viewer:, user_entity:, record_entity:)
      @viewer = viewer
      @user_entity = user_entity
      @record_entity = record_entity
    end
  end
end
