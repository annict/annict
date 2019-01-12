# frozen_string_literal: true

module Types
  module Enums
    class SeasonName < Types::Enums::Base
      description "Season name"

      value "WINTER", ""
      value "SPRING", ""
      value "SUMMER", ""
      value "AUTUMN", ""
    end
  end
end
