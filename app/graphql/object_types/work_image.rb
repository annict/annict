# frozen_string_literal: true

ObjectTypes::WorkImage = GraphQL::ObjectType.define do
  name "WorkImage"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :work, !ObjectTypes::Work do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :internal_url, types.String do
    argument :size, !types.String

    resolve ->(objs, args, ctx) {
      return nil unless ctx[:doorkeeper_token].owner.role.admin?
      obj = objs.first
      ann_image_url obj, :attachment, size: args[:size]
    }
  end

  field :facebook_og_image_url, types.String do
    resolve ->(objs, _args, _ctx) {
      obj = objs.first
      obj.work.facebook_og_image_url
    }
  end

  field :twitter_avatar_url, types.String do
    resolve ->(objs, _args, _ctx) {
      obj = objs.first
      obj.work.twitter_avatar_url
    }
  end
  %i(mini normal bigger).each do |size|
    field "twitter_#{size}_avatar_url".to_sym, types.String do
      resolve ->(objs, _args, _ctx) {
        obj = objs.first
        obj.work.twitter_avatar_url(size)
      }
    end
  end

  field :recommended_image_url, types.String do
    resolve ->(objs, _args, _ctx) {
      obj = objs.first
      obj.work.recommended_image_url
    }
  end
end
