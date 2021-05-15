# frozen_string_literal: true

module V4
  class AnimeRecordCardComponent < V4::ApplicationComponent
    def initialize(anime_entity:, anime_record_entity:)
      @anime_entity = anime_entity
      @anime_record_entity = anime_record_entity
    end
  end
end
