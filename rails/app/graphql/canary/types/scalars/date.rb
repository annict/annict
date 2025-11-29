# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Scalars
      class Date < Canary::Types::Scalars::Base
        def self.coerce_input(value, _ctx)
          Date.parse(value)
        end

        def self.coerce_result(value, _ctx)
          value.iso8601
        end
      end
    end
  end
end
