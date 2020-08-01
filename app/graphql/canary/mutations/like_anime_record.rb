# frozen_string_literal: true

module Canary
  module Mutations
    class LikeAnimeRecord < Canary::Mutations::Base
      argument :work_record_id, ID, required: true

      field :work_record, Canary::Types::Objects::AnimeRecordType, null: false

      def resolve(work_record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work_record = AnimeRecord.only_kept.find_by_graphql_id(work_record_id)
        context[:viewer].like(work_record)

        {
          work_record: work_record
        }
      end
    end
  end
end
