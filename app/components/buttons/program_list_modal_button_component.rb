# frozen_string_literal: true

module Buttons
  class ProgramListModalButtonComponent < ApplicationComponent
    def initialize(anime_id:, class_name: "")
      @anime_id = anime_id
      @class_name = class_name
    end
  end
end
