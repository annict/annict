# typed: false
# frozen_string_literal: true

module Canary
  class AnnictSchema < GraphQL::Schema
    query Canary::Types::Objects::Query
    mutation Canary::Types::Objects::Mutation

    use GraphQL::Batch
    use GraphQL::FragmentCache

    default_max_page_size 50

    def self.id_from_object(object, type_definition, query_ctx = nil)
      # https://github.com/rmosolgo/graphql-ruby/blob/e94569f51606f3635d8beb042d9f9c769dc9ae49/CHANGELOG.md#breaking-changes-3
      type_name = type_definition.respond_to?(:graphql_name) ? type_definition.graphql_name : type_definition.name

      GraphQL::Schema::UniqueWithinType.encode(type_name, object.id)
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
