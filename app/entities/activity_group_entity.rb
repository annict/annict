# frozen_string_literal: true

class ActivityGroupEntity < ApplicationEntity
  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :itemable_type, Types::ActivityResourceKinds
  attribute? :single, Types::Bool
  attribute? :activities_count, Types::Integer
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :itemables, Types::Array.of(EpisodeRecordEntity | StatusEntity | WorkRecordEntity)
  attribute? :activities_page_info, PageInfoEntity

  def status?
    itemable_type == "status"
  end

  def episode_record?
    itemable_type == "episode_record"
  end

  def work_record?
    itemable_type == "work_record"
  end
end
