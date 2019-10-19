# frozen_string_literal: true

module Canary
  class AnnictSchema < GraphQL::Schema
    query Canary::Types::Objects::Query
    mutation Canary::Types::Objects::Mutation

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
        Canary::Types::Unions::ActivityItem
      when EpisodeRecord
        Canary::Types::Objects::EpisodeRecordType
      when Episode
        Canary::Types::Objects::EpisodeType
      when MultipleEpisodeRecord
        Canary::Types::Objects::MultipleEpisodeRecordType
      when Organization
        Canary::Types::Objects::OrganizationType
      when Person
        Canary::Types::Objects::PersonType
      when Slot
        Canary::Types::Objects::SlotType
      when Program
        Canary::Types::Objects::ProgramType
      when Status
        Canary::Types::Objects::StatusType
      when User
        Canary::Types::Objects::UserType
      when Work
        Canary::Types::Objects::WorkType
      when WorkImage
        Canary::Types::Objects::WorkImageType
      when WorkRecord
        Canary::Types::Objects::WorkRecordType
      else
        raise "Unexpected object: #{obj}"
      end
    end
  end
end
