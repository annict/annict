# frozen_string_literal: true

module Mutations
  class CreateRecord < Mutations::Base
    include V4::GraphqlRunnable

    argument :episode_id, ID, required: true
    argument :comment, String, required: false
    argument :rating_state, Types::Enums::RatingState, required: false
    argument :share_twitter, Boolean, required: false
    argument :share_facebook, Boolean, required: false

    field :record, Types::Objects::RecordType, null: true

    def resolve(episode_id:, comment: nil, rating_state: nil, share_twitter: nil, share_facebook: nil)
      raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

      viewer = context[:viewer]
      episode = Episode.only_kept.find_by_graphql_id(episode_id)

      episode_record_entity, err = CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: viewer)
      ).create(
        episode: episode,
        params: {
          rating_state: rating_state,
          body: comment,
          share_to_twitter: share_twitter&.to_s
        }
      )

      if err
        raise GraphQL::ExecutionError, err.message
      end

      {
        record: viewer.episode_records.find(episode_record_entity.id)
      }
    end
  end
end
