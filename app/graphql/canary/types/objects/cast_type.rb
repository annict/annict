# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class CastType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer,
          null: false

        field :name, String,
          null: false,
          description: "出演者名"

        field :name_en, String,
          null: false,
          description: "出演者名 (英語)"

        field :local_accurated_name, String,
          null: false,
          description: "出演者名。出演当時と名前が異なる場合新旧2つの名前を併記する。例: 長島雄一 (チョー)"

        field :sort_number, Integer,
          null: false,
          description: "ソート番号"

        field :work, Canary::Types::Objects::WorkType,
          null: false

        field :character, Canary::Types::Objects::CharacterType,
          null: false

        field :person, Canary::Types::Objects::PersonType,
          null: false

        def local_accurated_name
          object.decorate.local_name_with_old
        end
      end
    end
  end
end
