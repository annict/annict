# frozen_string_literal: true

class AnimeRecordCardComponent < ApplicationComponent
  def initialize(anime_entity:, anime_record_entity:)
    @anime_entity = anime_entity
    @anime_record_entity = anime_record_entity
  end
end
