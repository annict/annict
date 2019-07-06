# frozen_string_literal: true

class AnnictSchema < GraphQL::Schema
  query Types::Objects::Query
  mutation Types::Objects::Mutation

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
      Types::Unions::ActivityItem
    when EpisodeRecord
      Types::Objects::RecordType
    when Episode
      Types::Objects::EpisodeType
    when MultipleEpisodeRecord
      Types::Objects::MultipleRecordType
    when Organization
      Types::Objects::OrganizationType
    when Person
      Types::Objects::PersonType
    when Status
      Types::Objects::StatusType
    when User
      Types::Objects::UserType
    when Work
      Types::Objects::WorkType
    when WorkImage
      Types::Objects::WorkImageType
    when WorkRecord
      Types::Objects::ReviewType
    else
      raise "Unexpected object: #{obj}"
    end
  end
end
