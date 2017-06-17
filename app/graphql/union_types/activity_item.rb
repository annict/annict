# frozen_string_literal: true

UnionTypes::ActivityItem = GraphQL::UnionType.define do
  name "ActivityItem"

  possible_types [
    ObjectTypes::Status,
    ObjectTypes::Record,
    ObjectTypes::MultipleRecord
  ]
end
