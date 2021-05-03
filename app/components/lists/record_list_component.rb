# frozen_string_literal: true

module Lists
  # TODO: RecordListComponent2 に置き換える
  class RecordListComponent < ApplicationComponent
    def initialize(viewer:, record_entities:, show_card: true)
      super
      @viewer = viewer
      @record_entities = record_entities
      @show_card = show_card
    end
  end
end
