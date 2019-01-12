# frozen_string_literal: true

module Types
  module Objects
    class PrefectureType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :name, String, null: false
    end
  end
end
