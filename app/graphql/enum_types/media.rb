# frozen_string_literal: true

EnumTypes::Media = GraphQL::EnumType.define do
  name "Media"
  description "Media of anime"

  value "TV", ""
  value "OVA", ""
  value "MOVIE", ""
  value "WEB", ""
  value "OTHER", ""
end
