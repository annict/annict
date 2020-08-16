# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class RecordItemableType < Types::Enums::Base
        graphql_name "RecordItemableType"

        value "ANIME_RECORD", ""
        value "EPISODE_RECORD", ""
      end
    end
  end
end
