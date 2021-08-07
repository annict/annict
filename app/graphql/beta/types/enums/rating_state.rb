# frozen_string_literal: true

module Beta
  module Types
    module Enums
      class RatingState < Beta::Types::Enums::Base
        value "GREAT", value: "great"
        value "GOOD", value: "good"
        value "AVERAGE", value: "average"
        value "BAD", value: "bad"
      end
    end
  end
end
