# frozen_string_literal: true

module Beta
  module Mutations
    class CreateRecord < Beta::Mutations::Base
      argument :episode_id, ID, required: true
      argument :comment, String, required: false
      argument :rating_state, Beta::Types::Enums::RatingState, required: false
      argument :share_twitter, Boolean, required: false
      argument :share_facebook, Boolean, required: false

      field :record, Beta::Types::Objects::RecordType, null: true

      def resolve(episode_id:, comment: nil, rating_state: nil, share_twitter: nil, share_facebook: nil)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:doorkeeper_token].writable?

        viewer = context[:viewer]
        episode = Episode.only_kept.find_by_graphql_id(episode_id)
        oauth_application = context[:doorkeeper_token].application

        form = Forms::EpisodeRecordForm.new(user: viewer, episode: episode, oauth_application: oauth_application)
        form.attributes = {
          comment: comment,
          rating: rating_state,
          share_to_twitter: share_twitter&.to_s
        }

        if form.invalid?
          raise GraphQL::ExecutionError, form.errors.full_messages.first
        end

        result = Creators::EpisodeRecordCreator.new(
          user: viewer,
          form: form
        ).call

        {
          record: viewer.episode_records.find_by!(record_id: result.record.id)
        }
      end
    end
  end
end
