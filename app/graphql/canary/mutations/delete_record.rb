# frozen_string_literal: true

module Canary
  module Mutations
    class DeleteRecord < Canary::Mutations::Base
      argument :record_id, ID, required: true

      field :user, Canary::Types::Objects::UserType, null: true

      def resolve(record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        record = context[:viewer].records.only_kept.find_by_graphql_id(record_id)
        Destroyers::RecordDestroyer.new(record: record).call

        {
          user: context[:viewer]
        }
      end
    end
  end
end
