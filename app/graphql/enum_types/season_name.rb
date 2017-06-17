# frozen_string_literal: true

EnumTypes::SeasonName = GraphQL::EnumType.define do
  name "SeasonName"
  description "Season name"

  value "WINTER", ""
  value "SPRING", ""
  value "SUMMER", ""
  value "AUTUMN", ""
end
