# frozen_string_literal: true

ObjectTypes::MultipleRecord = GraphQL::ObjectType.define do
  name "MultipleRecord"

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

  connection :records, ObjectTypes::Record.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(Checkin, :multiple_record_id).load([obj.id])
    }
  end

  field :createdAt, !ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.created_at
    }
  end
end
