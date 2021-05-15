# frozen_string_literal: true

module V4
  class AnimeRecordFooterComponent < V4::ApplicationComponent
    def initialize(record_entity:)
      @record_entity = record_entity
    end
  end
end
