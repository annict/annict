# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class SeasonType < Canary::Types::Enums::Base
        description "シーズン"

        value "WINTER", "冬"
        value "SPRING", "春"
        value "SUMMER", "夏"
        value "AUTUMN", "秋"
      end
    end
  end
end
