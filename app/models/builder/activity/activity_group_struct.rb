# frozen_string_literal: true

module Builder::Activity
  class ActivityGroupStruct < Builder::Activity::BaseStruct
    attribute :itemable_type, Types::ActivityItemableType
    attribute :activities_count, Types::Integer
    attribute :created_at, Types::Params::Time
    attribute :user, Types.Instance(User)
    attribute :itemables, Types::Array.of(Types.Instance(Record) | Types.Instance(Status))
  end
end
