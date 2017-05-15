# frozen_string_literal: true

Types::WorkImageType = GraphQL::ObjectType.define do
  name "WorkImage"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :internal_url, types.String do
    argument :size, !types.String

    resolve ->(obj, args, ctx) {
      return nil unless ctx[:doorkeeper_token].owner.role.admin?
      ann_image_url obj, :attachment, size: args[:size]
    }
  end

  field :facebook_og_image_url, types.String do
    resolve ->(obj, _args, _ctx) {
      obj.work.facebook_og_image_url
    }
  end

  field :twitter_avatar_url, types.String do
    resolve ->(obj, _args, _ctx) {
      obj.work.twitter_avatar_url
    }
  end
  %i(mini normal bigger).each do |size|
    field "twitter_#{size}_avatar_url".to_sym, types.String do
      resolve ->(obj, _args, _ctx) {
        obj.work.twitter_avatar_url(size)
      }
    end
  end

  field :recommended_image_url, types.String do
    resolve ->(obj, _args, _ctx) {
      obj.work.recommended_image_url
    }
  end
end
