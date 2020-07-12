# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class RecordItemableType < Types::Enums::Base
        graphql_name "RecordItemableType"

        value "EPISODE_RECORD", ""
        value "WORK_RECORD", ""
      end
    end
  end
end
