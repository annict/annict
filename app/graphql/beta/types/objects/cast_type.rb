# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class CastType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :name, String, null: false
        field :name_en, String, null: false
        field :sort_number, Integer, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :character, Beta::Types::Objects::CharacterType, null: false
        field :person, Beta::Types::Objects::PersonType, null: false
      end
    end
  end
end
