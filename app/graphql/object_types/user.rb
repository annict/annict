# frozen_string_literal: true

ObjectTypes::User = GraphQL::ObjectType.define do
  name "User"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  connection :activities, Connections::ActivityConnection do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(Activity, :user_id).load([obj.id])
    }
  end

  connection :followingActivities, Connections::ActivityConnection do
    resolve ->(obj, _args, _ctx) {
      following_ids = obj.followings.pluck(:id)
      following_ids << obj.id
      ForeignKeyLoader.for(Activity, :user_id).load(following_ids)
    }
  end

  connection :followers, ObjectTypes::User.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(User, :id).load(obj.followers.pluck(:id))
    }
  end

  connection :following, ObjectTypes::User.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(User, :id).load(obj.followings.pluck(:id))
    }
  end

  field :username, !types.String
  field :name, !types.String
  field :description, !types.String
  field :url, types.String

  field :avatarUrl, types.String do
    resolve ->(obj, _args, _ctx) {
      ann_api_assets_url(obj.profile, :tombo_avatar)
    }
  end

  field :records_count, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.checkins_count
    }
  end

  field :createdAt, !ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.created_at
    }
  end

  field :viewerCanFollow, !types.Boolean do
    resolve ->(obj, _args, ctx) {
      viewer = ctx[:doorkeeper_token].owner
      viewer != obj && !viewer.following?(obj)
    }
  end

  field :viewerIsFollowing, !types.Boolean do
    resolve ->(obj, _args, ctx) {
      viewer = ctx[:doorkeeper_token].owner
      viewer.following?(obj)
    }
  end

  field :email, types.String do
    resolve ->(obj, _args, ctx) {
      return nil if ctx[:doorkeeper_token].owner != obj
      obj.email
    }
  end

  field :notifications_count, types.Int do
    resolve ->(obj, _args, ctx) {
      return nil if ctx[:doorkeeper_token].owner != obj
      obj.notifications_count
    }
  end
end
