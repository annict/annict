# frozen_string_literal: true

ObjectTypes::WorkImage = GraphQL::ObjectType.define do
  name "WorkImage"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, types.Int do
    resolve ->(obj, _args, _ctx) {
      return nil if obj.blank?
      obj.id
    }
  end

  field :work, ObjectTypes::Work do
    resolve ->(obj, _args, _ctx) {
      return nil if obj.blank?
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :internalUrl, types.String do
    argument :size, !types.String

    resolve ->(obj, args, ctx) {
      return unless ctx[:internal]
      return unless ctx[:oauth_application]&.official?
      return if obj.blank?
      ann_image_url obj, :attachment, size: args[:size]
    }
  end

  field :facebookOgImageUrl, types.String do
    resolve ->(obj, _args, _ctx) {
      return "" if obj.blank?
      obj.work.facebook_og_image_url
    }
  end

  field :twitterAvatarUrl, types.String do
    resolve ->(obj, _args, _ctx) {
      return "" if obj.blank?
      obj.work.twitter_avatar_url
    }
  end
  %i(mini normal bigger).each do |size|
    field "twitter#{size.capitalize}AvatarUrl".to_sym, types.String do
      resolve ->(obj, _args, _ctx) {
        return "" if obj.blank?
        obj.work.twitter_avatar_url(size)
      }
    end
  end

  field :recommendedImageUrl, types.String do
    resolve ->(obj, _args, _ctx) {
      return "" if obj.blank?
      obj.work.recommended_image_url
    }
  end
end
