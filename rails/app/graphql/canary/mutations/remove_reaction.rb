# typed: false
# frozen_string_literal: true

module Canary
  module Mutations
    class RemoveReaction < Canary::Mutations::Base
      argument :reactable_id, ID, required: true
      argument :content, Canary::Types::Enums::ReactionContent, required: true

      field :reaction, Canary::Types::Objects::ReactionType, null: true
      field :reactable, Canary::Types::Interfaces::Reactable, null: true

      def resolve(reactable_id:, content:)
        raise Annict::Errors::InvalidAPITokenScopeError unless context[:writable]

        viewer = context[:viewer]
        reactable = Canary::AnnictSchema.object_from_id(reactable_id)

        return if reactable.nil?

        object_type = Canary::AnnictSchema.resolve_type(nil, reactable, nil)

        unless object_type.interfaces.include?(Canary::Types::Interfaces::Reactable)
          raise GraphQL::ExecutionError, "reactableId does not implement Reactable interface"
        end

        recipient = case reactable
        when Record
          reactable.episode_record.presence || reactable.work_record
        else
          reactable
        end

        like = viewer.unlike(recipient)

        {
          reaction: like,
          reactable: reactable
        }
      end
    end
  end
end
