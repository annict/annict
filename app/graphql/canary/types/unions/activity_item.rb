# frozen_string_literal: true

module Types
  module Unions
    class ActivityItem < Types::Unions::Base
      possible_types Types::Objects::StatusType,
        Types::Objects::RecordType,
        Types::Objects::ReviewType,
        Types::Objects::MultipleRecordType
    end
  end
end
