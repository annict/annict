# frozen_string_literal: true

module Types
  module Objects
    class CastType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :name, String, null: false
      field :name_en, String, null: false
      field :sort_number, Integer, null: false
      field :work, Types::Objects::WorkType, null: false
      field :character, Types::Objects::CharacterType, null: false
      field :person, Types::Objects::PersonType, null: false
    end
  end
end
