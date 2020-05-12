# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class ActivityItem < Canary::Types::Unions::Base
        possible_types Canary::Types::Objects::StatusType,
          Canary::Types::Objects::EpisodeRecordType,
          Canary::Types::Objects::WorkRecordType
      end
    end
  end
end
