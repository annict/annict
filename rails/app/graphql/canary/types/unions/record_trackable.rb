# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class RecordTrackable < Canary::Types::Unions::Base
        graphql_name "RecordTrackable"

        possible_types Canary::Types::Objects::WorkType, Canary::Types::Objects::EpisodeType
      end
    end
  end
end
