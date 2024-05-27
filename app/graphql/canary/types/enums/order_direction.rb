# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class OrderDirection < Canary::Types::Enums::Base
        value "ASC", "昇順"
        value "DESC", "降順"
      end
    end
  end
end
