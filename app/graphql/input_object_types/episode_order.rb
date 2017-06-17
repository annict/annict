# frozen_string_literal: true

InputObjectTypes::EpisodeOrder = GraphQL::InputObjectType.define do
  name "EpisodeOrder"

  argument :field, !EnumTypes::EpisodeOrderField
  argument :direction, !EnumTypes::OrderDirection
end
