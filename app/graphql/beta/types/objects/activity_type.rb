# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class ActivityType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        field :annict_id, Integer, null: false
        field :user, Beta::Types::Objects::UserType, null: false

        def user
          Beta::RecordLoader.for(User).load(object.user_id)
        end
      end
    end
  end
end
