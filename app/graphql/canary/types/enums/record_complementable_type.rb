# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class RecordComplementableType < Types::Enums::Base
        graphql_name "RecordComplementableType"

        value "ANIME_RECORD", ""
        value "EPISODE_RECORD", ""
      end
    end
  end
end
