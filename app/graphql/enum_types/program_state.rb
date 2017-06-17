# frozen_string_literal: true

EnumTypes::ProgramState = GraphQL::EnumType.define do
  name "ProgramState"

  value "PUBLISHED", ""
  value "HIDDEN", ""
end
