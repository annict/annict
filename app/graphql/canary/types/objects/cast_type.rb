# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class CastType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer,
          null: false

        field :name, String,
          null: false,
          description: "出演者名"

        field :name_en, String,
          null: false,
          description: "出演者名 (英語)"

        field :accurate_name, String,
          null: false,
          description: "出演者名。出演当時と名前が異なる場合新旧2つの名前を併記する。例: 長島雄一 (チョー)"

        field :accurate_name_en, String,
          null: false,
          description: "出演者名 (英)。出演当時と名前が異なる場合新旧2つの名前を併記する。例: Nagashima, Yuuichi (Cho)"

        field :sort_number, Integer,
          null: false,
          description: "ソート番号"

        field :work, Canary::Types::Objects::WorkType,
          null: false

        field :character, Canary::Types::Objects::CharacterType,
          null: false

        field :person, Canary::Types::Objects::PersonType,
          null: false

        def accurate_name
          object.decorate.accurate_name
        end

        def accurate_name_en
          object.decorate.accurate_name_en
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end
      end
    end
  end
end
