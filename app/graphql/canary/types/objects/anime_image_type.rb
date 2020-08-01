# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class AnimeImageType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: true
        field :work, Canary::Types::Objects::AnimeType, null: true
        field :facebook_og_image_url, String, null: true
        field :twitter_avatar_url, String, null: true
        field :twitter_mini_avatar_url, String, null: true
        field :twitter_normal_avatar_url, String, null: true
        field :twitter_bigger_avatar_url, String, null: true
        field :recommended_image_url, String, null: true

        field :internal_url, String,
          null: false,
          description: "このフィールドの値は公開されていません" do
          argument :size, String, required: true
        end

        def work
          RecordLoader.for(Anime).load(object.anime_id)
        end

        def internal_url(size:)
          return "" unless context[:admin]

          ann_image_url object, :image, size: size
        end

        def facebook_og_image_url
          return "" if object.blank?

          object.work.facebook_og_image_url
        end

        def twitter_avatar_url
          return "" if object.blank?

          object.work.twitter_avatar_url
        end

        def twitter_mini_avatar_url
          return "" if object.blank?

          object.work.twitter_avatar_url(:mini)
        end

        def twitter_normal_avatar_url
          return "" if object.blank?

          object.work.twitter_avatar_url(:normal)
        end

        def twitter_bigger_avatar_url
          return "" if object.blank?

          object.work.twitter_avatar_url(:bigger)
        end

        def recommended_image_url
          return "" if object.blank?

          object.work.recommended_image_url
        end
      end
    end
  end
end
