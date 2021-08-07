# frozen_string_literal: true

module Beta
  module Types
    module Enums
      class SeasonName < Beta::Types::Enums::Base
        description "Season name"

        value "WINTER", ""
        value "SPRING", ""
        value "SUMMER", ""
        value "AUTUMN", ""
      end
    end
  end
end
