# frozen_string_literal: true

class AnnictSchema < GraphQL::Schema
  query ObjectTypes::Query
  mutation ObjectTypes::Mutation

  use GraphQL::Batch

  def self.id_from_object(object, type_definition, _query_ctx)
    GraphQL::Schema::UniqueWithinType.encode(type_definition.name, object.id)
  end

  def self.object_from_id(id, _query_ctx)
    type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(id)

    return nil if type_name.blank? || item_id.blank?

    Object.const_get(type_name).find(item_id)
  end

  def self.resolve_type(_type, obj, _ctx)
    case obj
    when Activity
      UnionTypes::Activity
    when EpisodeRecord
      ObjectTypes::Record
    when Episode
      ObjectTypes::Episode
    when MultipleEpisodeRecord
      ObjectTypes::MultipleRecord
    when WorkRecord
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
  end
end
