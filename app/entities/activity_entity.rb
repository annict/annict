# frozen_string_literal: true

class ActivityEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :resource_type, Types::ActivityResourceKinds
  attribute? :resources_count, Types::Integer
  attribute? :single, Types::Bool
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :resources, Types::Array.of(EpisodeRecordEntity | StatusEntity | WorkRecordEntity)
end
