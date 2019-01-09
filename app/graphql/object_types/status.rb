# frozen_string_literal: true

ObjectTypes::Status = GraphQL::ObjectType.define do
  name "Status"

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

  field :state, !Types::Enum::StatusState do
    resolve ->(obj, _args, _ctx) {
      obj.kind.upcase
    }
  end

  field :likesCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.likes_count
    }
  end

  field :createdAt, !ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.created_at
    }
  end
end
