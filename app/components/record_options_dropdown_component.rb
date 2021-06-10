# frozen_string_literal: true

class RecordOptionsDropdownComponent < ApplicationComponent
  def initialize(viewer:, user_entity:, record_entity:)
    @viewer = viewer
    @user_entity = user_entity
    @record_entity = record_entity
  end
end
