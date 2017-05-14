# frozen_string_literal: true

Types::EpisodeType = GraphQL::ObjectType.define do
  name "Episode"
  description "An episode of a work"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :title, !types.String
  # field :prev_episode, types[Types::EpisodeType]
end
