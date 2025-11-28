# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class WorkImageType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: true
        field :work, Canary::Types::Objects::WorkType, null: true
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
          RecordLoader.for(Work).load(object.work_id)
        end

        def internal_url(size:)
          return "" unless context[:admin]

          width, height = size.split("x")

          v6_ann_image_url object, :image, format: "jpg", height: height, width: width
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
