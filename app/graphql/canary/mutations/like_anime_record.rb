# frozen_string_literal: true

module Canary
  module Mutations
    class LikeAnimeRecord < Canary::Mutations::Base
      argument :anime_record_id, ID, required: true

      field :anime_record, Canary::Types::Objects::AnimeRecordType, null: false

      def resolve(anime_record_id:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        work_record = WorkRecord.only_kept.find_by_graphql_id(anime_record_id)
        context[:viewer].like(work_record)

        {
          anime_record: work_record
        }
      end
    end
  end
end
