# frozen_string_literal: true

module Canary
  module Mutations
    class UpdateStatus < Canary::Mutations::Base
      argument :anime_id, ID, required: true
      argument :kind, Canary::Types::Enums::StatusKind, required: true

      field :anime, Canary::Types::Objects::AnimeType, null: true

      def resolve(anime_id:, kind:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        anime = Anime.only_kept.find_by_graphql_id(anime_id)

        form = Forms::StatusForm.new(anime: anime, kind: kind)

        if form.invalid?
          raise GraphQL::ExecutionError, "status update failed"
        end

        Updaters::StatusUpdater.new(user: viewer, form: form).call

        {
          anime: anime
        }
      end
    end
  end
end
