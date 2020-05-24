# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class ActivityItemableType < Types::Enums::Base
        graphql_name "ActivityItemableType"

        value "EPISODE_RECORD", ""
        value "STATUS", ""
        value "WORK_RECORD", ""
      end
    end
  end
end
