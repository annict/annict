# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class ActivityItemable < Canary::Types::Unions::Base
        graphql_name "ActivityItemable"

        possible_types Canary::Types::Objects::StatusType,
          Canary::Types::Objects::EpisodeRecordType,
          Canary::Types::Objects::AnimeRecordType
      end
    end
  end
end
