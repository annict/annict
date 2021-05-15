# frozen_string_literal: true

module Deprecated
  class AnimeRecordFooterComponent < Deprecated::ApplicationComponent
    def initialize(record_entity:)
      @record_entity = record_entity
    end
  end
end
