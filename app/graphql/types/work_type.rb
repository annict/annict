# frozen_string_literal: true

Types::WorkType = GraphQL::ObjectType.define do
  name "Work"
  description "An anime title"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  connection :episodes, Types::EpisodeType.connection_type

  field :title, !types.String

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  field :image, Types::WorkImageType do
    resolve ->(obj, _args, _ctx) {
      obj.work_image
    }
  end
end
