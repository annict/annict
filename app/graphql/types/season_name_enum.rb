# frozen_string_literal: true

Types::SeasonNameEnum = GraphQL::EnumType.define do
  name "SeasonName"
  description "Season name"

  value "WINTER", ""
  value "SPRING", ""
  value "SUMMER", ""
  value "AUTUMN", ""
end
