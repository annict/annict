# frozen_string_literal: true

module Types
  module Objects
    class ActivityType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      field :annict_id, Integer, null: false
      field :user, Types::Objects::UserType, null: false

      def user
        RecordLoader.for(User).load(object.user_id)
      end
    end
  end
end
