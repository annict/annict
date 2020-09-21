# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class RecordComplementable < Canary::Types::Unions::Base
        graphql_name "RecordComplementable"

        possible_types Canary::Types::Objects::EpisodeRecordType, Canary::Types::Objects::AnimeRecordType
      end
    end
  end
end
