# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class ActivityType < Types::Enums::Base
        graphql_name "ActivityType"

        value "EPISODE_RECORD", ""
        value "STATUS", ""
        value "WORK_RECORD", ""
      end
    end
  end
end
