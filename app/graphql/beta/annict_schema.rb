# frozen_string_literal: true

module Beta
  class AnnictSchema < GraphQL::Schema
    RENAMED_TYPE_MAPPING = {
      "Work" => "Anime",
      "WorkImage" => "AnimeImage",
      "WorkRecord" => "AnimeRecord"
    }.freeze

    query Beta::Types::Objects::Query
    mutation Beta::Types::Objects::Mutation

    use GraphQL::Batch

    def self.id_from_object(object, type_definition, query_ctx = nil)
      type_name = RENAMED_TYPE_MAPPING.invert[type_definition.name].presence || type_definition.name
      GraphQL::Schema::UniqueWithinType.encode(type_name, object.id)
    end

    def self.object_from_id(id, _query_ctx)
      type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(id)
      type_name = RENAMED_TYPE_MAPPING[type_name].presence || type_name

      return nil if type_name.blank? || item_id.blank?

      Object.const_get(type_name).find(item_id)
    end

    def self.resolve_type(_type, obj, _ctx)
      case obj
      when Activity
        Beta::Types::Unions::ActivityItem
      when EpisodeRecord
        Beta::Types::Objects::RecordType
      when Episode
        Beta::Types::Objects::EpisodeType
      when MultipleEpisodeRecord
        Beta::Types::Objects::MultipleRecordType
      when Organization
        Beta::Types::Objects::OrganizationType
      when Person
        Beta::Types::Objects::PersonType
      when Status
        Beta::Types::Objects::StatusType
      when User
        Beta::Types::Objects::UserType
      when Anime
        Beta::Types::Objects::WorkType
      when AnimeImage
        Beta::Types::Objects::WorkImageType
      when AnimeRecord
        Beta::Types::Objects::ReviewType
      else
        raise "Unexpected object: #{obj}"
      end
    end
  end
end
