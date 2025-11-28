# typed: false
# frozen_string_literal: true

module Canary
  module Resolvers
    class Record < Canary::Resolvers::Base
      def resolve(database_id:)
        object.records.only_kept.find_by(id: database_id)
      end
    end
  end
end
