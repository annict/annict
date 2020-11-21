# frozen_string_literal: true

module Dropdowns
  class AnimeMenuDropdownComponent < ApplicationComponent
    def initialize(anime_id:, class_name: "")
      @anime_id = anime_id
      @class_name = class_name
    end
  end
end
