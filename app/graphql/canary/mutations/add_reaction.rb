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

        recipient = case reactable
        when Record
          if reactable.episode_record?
            reactable.episode_record
          else
            reactable.work_record
          end
        end

        like = viewer.likes.find_by(recipient: recipient)

        if like
          return {
            reaction: like,
            reactable: reactable
          }
        end

        like = viewer.like(recipient)

        if recipient.is_a?(EpisodeRecord)
          EmailNotificationService.send_email(
            "liked_episode_record",
            reactable.user,
            viewer.id,
            recipient.id
          )
        end

        {
          reaction: like,
          reactable: reactable
        }
      end
    end
  end
end
