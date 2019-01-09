# frozen_string_literal: true

InputObjectTypes::EpisodeOrder = GraphQL::InputObjectType.define do
  name "EpisodeOrder"

  argument :field, !Types::Enum::EpisodeOrderField
  argument :direction, !Types::Enum::OrderDirection
end
