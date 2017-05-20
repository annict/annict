# frozen_string_literal: true

EnumTypes::StatusState = GraphQL::EnumType.define do
  name "StatusState"

  value "WANNA_WATCH", ""
  value "WATCHING", ""
  value "WATCHED", ""
  value "ON_HOLD", ""
  value "STOP_WATCHING", ""
end
