# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class CastType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :name, String, null: false,
          description: "役名"
        field :name_en, String, null: false,
          description: "役名 (英語)"
        field :sort_number, Integer, null: false,
          description: "ソート番号"
        field :work, Canary::Types::Objects::WorkType, null: false
        field :character, Canary::Types::Objects::CharacterType, null: false
        field :person, Canary::Types::Objects::PersonType, null: false
      end
    end
  end
end
