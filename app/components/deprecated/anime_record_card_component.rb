# frozen_string_literal: true

module Deprecated
  class AnimeRecordCardComponent < Deprecated::ApplicationComponent
    def initialize(anime_entity:, anime_record_entity:)
      @anime_entity = anime_entity
      @anime_record_entity = anime_record_entity
    end
  end
end
