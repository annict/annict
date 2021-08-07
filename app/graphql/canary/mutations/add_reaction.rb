# frozen_string_literal: true

module Canary
  module Mutations
    class AddReaction < Canary::Mutations::Base
      argument :reactable_id, ID, required: true
      argument :content, Canary::Types::Enums::ReactionContent, required: true

      field :reaction, Canary::Types::Objects::ReactionType, null: false
      field :reactable, Canary::Types::Interfaces::Reactable, null: false

      def resolve(reactable_id:, content:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        reactable = Canary::AnnictSchema.object_from_id(reactable_id)

        result = V4::AddReactionService.new(user: viewer, reactable: reactable).call

        {
          reaction: result.reaction,
          reactable: reactable
        }
      end
    end
  end
end
