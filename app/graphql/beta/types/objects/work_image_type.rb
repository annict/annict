# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class WorkImageType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: true
        field :work, Beta::Types::Objects::WorkType, null: true
        field :facebook_og_image_url, String, null: true
        field :twitter_avatar_url, String, null: true
        field :twitter_mini_avatar_url, String, null: true
        field :twitter_normal_avatar_url, String, null: true
        field :twitter_bigger_avatar_url, String, null: true
        field :recommended_image_url, String, null: true
        field :copyright, String, null: true

        field :internal_url, String, null: true do
          argument :size, String, required: true
        end

        def work
          Beta::RecordLoader.for(Work).load(object.work_id)
        end

        def internal_url(size:)
          return unless context[:doorkeeper_token].owner.role.admin?
          return "" if object.blank?

          ann_work_image_url object, width: size.split(":")[0]
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
