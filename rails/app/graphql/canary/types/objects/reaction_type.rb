# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ReactionType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        field :user, Canary::Types::Objects::UserType, null: false
        field :content, Canary::Types::Enums::ReactionContent, null: false
        field :reactable, Canary::Types::Interfaces::Reactable, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false

        def content
          :heart
        end

        def reactable
          case object.recipient
          when EpisodeRecord, WorkRecord
            object.recipient.record
          else
            object.recipient
          end
        end
      end
    end
  end
end
