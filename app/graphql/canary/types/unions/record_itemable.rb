# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class RecordItemable < Canary::Types::Unions::Base
        graphql_name "RecordItemable"

        possible_types Canary::Types::Objects::EpisodeRecordType, Canary::Types::Objects::WorkRecordType
      end
    end
  end
end
