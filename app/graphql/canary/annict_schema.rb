# frozen_string_literal: true

module Canary
  class AnnictSchema < GraphQL::Schema
    query Canary::Types::Objects::Query
    mutation Canary::Types::Objects::Mutation

    use GraphQL::Batch
    use GraphQL::FragmentCache

    default_max_page_size 50

    def self.id_from_object(object, type_definition, query_ctx = nil)
      GraphQL::Schema::UniqueWithinType.encode(type_definition.name, object.id)
    end

    def self.object_from_id(id, query_ctx = nil)
      type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(id)

      return nil if type_name.blank? || item_id.blank?

      Object.const_get(type_name).find(item_id)
    end

    def self.resolve_type(_type, obj, _ctx)
      case obj
      when Activity
        Canary::Types::Objects::ActivityType
      when ActivityGroup
        Canary::Types::Objects::ActivityGroupType
      when Anime
        Canary::Types::Objects::AnimeType
      when EpisodeRecord
        Canary::Types::Objects::EpisodeRecordType
      when Episode
        Canary::Types::Objects::EpisodeType
      when Organization
        Canary::Types::Objects::OrganizationType
      when Person
        Canary::Types::Objects::PersonType
      when Slot
        Canary::Types::Objects::SlotType
      when Program
        Canary::Types::Objects::ProgramType
      when Record
        Canary::Types::Objects::RecordType
      when Status
        Canary::Types::Objects::StatusType
      when User
        Canary::Types::Objects::UserType
      when AnimeImage
        Canary::Types::Objects::AnimeImageType
      when AnimeRecord
        Canary::Types::Objects::AnimeRecordType
      else
        raise "Unexpected object: #{obj}"
      end
    end
  end
end
