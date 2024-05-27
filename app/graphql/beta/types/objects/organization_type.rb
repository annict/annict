# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class OrganizationType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :name, String, null: false
        field :name_kana, String, null: false
        field :name_en, String, null: false
        field :url, String, null: false
        field :url_en, String, null: false
        field :wikipedia_url, String, null: false
        field :wikipedia_url_en, String, null: false
        field :twitter_username, String, null: false
        field :twitter_username_en, String, null: false
        field :favorite_organizations_count, Integer, null: false
        field :staffs_count, Integer, null: false

        def favorite_organizations_count
          object.favorite_users_count
        end
      end
    end
  end
end
