# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class Recordable < Canary::Types::Unions::Base
        graphql_name "Recordable"

        possible_types Canary::Types::Objects::EpisodeRecordType, Canary::Types::Objects::AnimeRecordType
      end
    end
  end
end
