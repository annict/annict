# frozen_string_literal: true

module ButtonGroups
  class AnimeButtonGroupComponent < ApplicationComponent
    def initialize(anime_id:)
      @anime_id = anime_id
    end
  end
end
