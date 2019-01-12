# frozen_string_literal: true

module Types
  module Objects
    class OrganizationType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

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
    end
  end
end
