# frozen_string_literal: true

module Beta
  class AnnictSchema < GraphQL::Schema
    query Beta::Types::Objects::Query
    mutation Beta::Types::Objects::Mutation

    use GraphQL::Batch

    def self.id_from_object(object, type_definition, query_ctx = nil)
      # https://github.com/rmosolgo/graphql-ruby/blob/e94569f51606f3635d8beb042d9f9c769dc9ae49/CHANGELOG.md#breaking-changes-3
      type_name = type_definition.respond_to?(:graphql_name) ? type_definition.graphql_name : type_definition.name

      GraphQL::Schema::UniqueWithinType.encode(type_name, object.id)
    end

    def self.decode_id(id, ctx: nil)
      type_name, item_id = GraphQL::Schema::UniqueWithinType.decode(id)

      if type_name.blank? || item_id.blank?
        raise "Unexpected id: #{id.inspect}"
      end

      new_type_name = case type_name
      when "Record" then "EpisodeRecord"
      when "Review" then "WorkRecord"
      else
        type_name
      end

      [new_type_name, item_id]
    end

    def self.object_from_id(id, _query_ctx)
      type_name, item_id = decode_id(id)

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
      when Work
        Beta::Types::Objects::WorkType
      when WorkImage
        Beta::Types::Objects::WorkImageType
      when WorkRecord
        Beta::Types::Objects::ReviewType
      else
        raise "Unexpected object: #{obj}"
      end
    end
  end
end
