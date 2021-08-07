# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class PersonType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :name, String, null: false
        field :name_kana, String, null: false
        field :name_en, String, null: false
        field :nickname, String, null: false
        field :nickname_en, String, null: false
        field :gender_text, String, null: false
        field :url, String, null: false
        field :url_en, String, null: false
        field :wikipedia_url, String, null: false
        field :wikipedia_url_en, String, null: false
        field :twitter_username, String, null: false
        field :twitter_username_en, String, null: false
        field :birthday, String, null: false
        field :blood_type, String, null: false
        field :height, String, null: false
        field :favorite_people_count, Integer, null: false
        field :casts_count, Integer, null: false
        field :staffs_count, Integer, null: false
        field :prefecture, Beta::Types::Objects::PrefectureType, null: false

        def favorite_people_count
          object.favorite_users_count
        end
      end
    end
  end
end
