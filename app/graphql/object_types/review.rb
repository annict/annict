# frozen_string_literal: true

ObjectTypes::Review = GraphQL::ObjectType.define do
  name "Review"

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

  field :title, types.String do
    resolve ->(obj, _args, _ctx) {
      obj.title.presence || I18n.t("noun.record_of_work", work_title: obj.work.decorate.local_title)
    }
  end

  field :body, !types.String

  WorkRecord::STATES.each do |state|
    field state.to_s.camelcase(:lower).to_sym, Types::Enum::RatingState do
      resolve ->(obj, _args, _ctx) {
        obj.send(state)&.upcase
      }
    end
  end

  field :likesCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.likes_count
    }
  end

  field :impressionsCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.impressions_count
    }
  end

  field :modifiedAt, ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.modified_at
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
