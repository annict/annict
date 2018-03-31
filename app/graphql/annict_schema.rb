# frozen_string_literal: true

AnnictSchema = GraphQL::Schema.define do
  query ObjectTypes::Query
  mutation ObjectTypes::Mutation

  use GraphQL::Batch

  id_from_object ->(object, type_definition, _query_ctx) {
    GraphQL::Schema::UniqueWithinType.encode(type_definition.name, object.id)
  }

  object_from_id ->(id, _query_ctx) {
    type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(id)

    return nil if type_name.blank? || item_id.blank?

    Object.const_get(type_name).find(item_id)
  }

  resolve_type ->(_type, obj, _ctx) {
    case obj
    when Activity
      UnionTypes::Activity
    when Record
      ObjectTypes::Record
    when Episode
      ObjectTypes::Episode
    when MultipleRecord
      ObjectTypes::MultipleRecord
    when Review
      ObjectTypes::Review
    when Status
      ObjectTypes::Status
    when User
      ObjectTypes::User
    when Work
      ObjectTypes::Work
    when WorkImage
      ObjectTypes::WorkImage
    else
      raise "Unexpected object: #{obj}"
    end
  }
end
