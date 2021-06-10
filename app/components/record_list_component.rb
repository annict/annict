# frozen_string_literal: true

class RecordListComponent < ApplicationComponent
  def initialize(viewer:, record_entities:, show_card: true)
    super
    @viewer = viewer
    @record_entities = record_entities
    @show_card = show_card
  end
end
