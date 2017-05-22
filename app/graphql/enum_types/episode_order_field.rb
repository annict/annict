# frozen_string_literal: true

EnumTypes::EpisodeOrderField = GraphQL::EnumType.define do
  name "EpisodeOrderField"

  value "CREATED_AT", ""
  value "SORT_NUMBER", ""
end
