# frozen_string_literal: true

# TODO: RecordFooterComponent に置き換える
class AnimeRecordFooterComponent < ApplicationComponent
  def initialize(record_entity:)
    @record_entity = record_entity
  end
end
