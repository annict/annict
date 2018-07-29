# frozen_string_literal: true

module ObjectTypes
  class Activity < Types::BaseObject
    implements GraphQL::Relay::Node.interface

    field :annict_id, Integer, null: false
    field :user, ObjectTypes::User, null: false

    def annict_id
      object.id
    end

    def user
      RecordLoader.for(User).load(object.user_id)
    end
  end
end
