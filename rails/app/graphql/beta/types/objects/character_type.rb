# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class CharacterType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :name, String, null: false
        field :name_kana, String, null: false
        field :name_en, String, null: false
        field :nickname, String, null: false
        field :nickname_en, String, null: false
        field :birthday, String, null: false
        field :birthday_en, String, null: false
        field :age, String, null: false
        field :age_en, String, null: false
        field :blood_type, String, null: false
        field :blood_type_en, String, null: false
        field :height, String, null: false
        field :height_en, String, null: false
        field :weight, String, null: false
        field :weight_en, String, null: false
        field :nationality, String, null: false
        field :nationality_en, String, null: false
        field :occupation, String, null: false
        field :occupation_en, String, null: false
        field :description, String, null: false
        field :description_en, String, null: false
        field :description_source, String, null: false
        field :description_source_en, String, null: false
        field :favorite_characters_count, Integer, null: false
        field :series, Beta::Types::Objects::SeriesType, null: true

        def favorite_characters_count
          object.favorite_users_count
        end
      end
    end
  end
end
