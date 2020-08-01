# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class TrailerType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :url, String, null: false
        field :title, String, null: false
        field :title_en, String, null: false
        field :sort_number, Integer, null: false
        field :is_youtube, Boolean, null: false
        field :work, Canary::Types::Objects::AnimeType, null: false

        field :internal_image_url, String, null: true, description: "このフィールドの値は公開されていません" do
          argument :size, String, required: true
        end

        def is_youtube
          object.youtube?
        end

        def work
          RecordLoader.for(Anime).load(object.work_id)
        end

        def internal_image_url(size:)
          return unless context[:admin]
          ann_image_url object, :image, size: size
        end
      end
    end
  end
end
