# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class ActivityItemableType < Types::Enums::Base
        graphql_name "ActivityItemableType"

        value "RECORD", ""
        value "STATUS", ""
      end
    end
  end
end
