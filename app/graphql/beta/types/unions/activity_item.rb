# frozen_string_literal: true

module Beta
  module Types
    module Unions
      class ActivityItem < Beta::Types::Unions::Base
        possible_types Beta::Types::Objects::StatusType,
          Beta::Types::Objects::RecordType,
          Beta::Types::Objects::ReviewType,
          Beta::Types::Objects::MultipleRecordType
      end
    end
  end
end
