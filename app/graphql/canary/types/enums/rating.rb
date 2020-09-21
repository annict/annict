# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class Rating < Canary::Types::Enums::Base
        value "GREAT", value: "great", description: "とても良い"
        value "GOOD", value: "good", description: "良い"
        value "AVERAGE", value: "average", description: "普通"
        value "BAD", value: "bad", description: "良くない"
      end
    end
  end
end
