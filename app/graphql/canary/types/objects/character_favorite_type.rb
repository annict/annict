# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class CharacterFavoriteType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        field :user, Canary::Types::Objects::UserType, null: false
        field :character, Canary::Types::Objects::CharacterType, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false

        def user
          Canary::RecordLoader.for(User).load(object.user_id)
        end

        def character
          Canary::RecordLoader.for(Character).load(object.character_id)
        end
      end
    end
  end
end
