# frozen_string_literal: true

class AnimeRecordFooterComponent < ApplicationComponent
  def initialize(record_entity:)
    @record_entity = record_entity
  end
end
