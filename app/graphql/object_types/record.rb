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

  field :rating, types.Float do
    resolve ->(obj, _args, _ctx) {
      obj.rating
    }
  end

  field :ratingState, Types::Enum::RatingState do
    resolve ->(obj, _args, _ctx) {
      obj.rating_state&.upcase
    }
  end

  field :modified, !types.Boolean do
    resolve ->(obj, _args, _ctx) {
      obj.modify_comment?
    }
  end

  field :likesCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.likes_count
    }
  end
  field :commentsCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.comments_count
    }
  end

  field :twitterClickCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.twitter_click_count
    }
  end
  field :facebookClickCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.facebook_click_count
    }
  end

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
