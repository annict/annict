# frozen_string_literal: true

ObjectTypes::Record = GraphQL::ObjectType.define do
  name "Record"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  field :user, !ObjectTypes::User do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(User).load(obj.user_id)
    }
  end

  field :work, !ObjectTypes::Work do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :episode, !ObjectTypes::Episode do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.episode_id)
    }
  end

  field :comment, types.String

  field :rating, EnumTypes::RatingState do
    resolve ->(obj, _args, _ctx) {
      return nil if obj.rating.blank?
      obj.rating > 3 ? "GOOD" : "BAD"
    }
  end

  field :modified, !types.Boolean do
    resolve ->(obj, _args, _ctx) {
      obj.modify_comment?
    }
  end

  field :likes_count, !types.Int
  field :comments_count, !types.Int
  field :twitter_click_count, !types.Int
  field :facebook_click_count, !types.Int

  field :createdAt, !ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.created_at
    }
  end

  field :updatedAt, !ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.updated_at
    }
  end
end
