# frozen_string_literal: true

class ActivityGroupEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :resource_type, Types::ActivityResourceKinds
  attribute? :single, Types::Bool
  attribute? :activities_count, Types::Integer
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :resources, Types::Array.of(EpisodeRecordEntity | StatusEntity | WorkRecordEntity)
end
