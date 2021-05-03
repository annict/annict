# frozen_string_literal: true

# TODO: RecordHeaderComponent2 に置き換える
class RecordHeaderComponent < ApplicationComponent
  def initialize(viewer:, user_entity:, record_entity:)
    @viewer = viewer
    @user_entity = user_entity
    @record_entity = record_entity
  end
end
