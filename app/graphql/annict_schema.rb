# frozen_string_literal: true

AnnictSchema = GraphQL::Schema.define do
  query Types::QueryType

  id_from_object ->(object, type_definition, _query_ctx) {
    GraphQL::Schema::UniqueWithinType.encode(type_definition.name, object.id)
  }

  object_from_id ->(id, _query_ctx) {
    type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(id)
    Object.const_get(type_name).find(item_id)
  }

  resolve_type ->(obj, _ctx) {
    case obj
    when Episode
      Types::EpisodeType
    when Work
      Types::WorkType
    when WorkImage
      Types::WorkImageType
    else
      raise "Unexpected object: #{obj}"
    end
  }
end
