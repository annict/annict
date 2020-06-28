# frozen_string_literal: true

class RecordEntity < ApplicationEntity
  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :itemable_type, Types::RecordResourceKinds
  attribute? :modified_at, Types::Params::Time.optional
  attribute? :created_at, Types::Params::Time
  attribute? :work, WorkEntity
  attribute? :itemable, EpisodeRecordEntity | WorkRecordEntity

  def self.from_node(record_node)
    attrs = {}

    if database_id = record_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if itemable_type = record_node["itemableType"]
      attrs[:itemable_type] = itemable_type.downcase
    end

    if modified_at = record_node["modifiedAt"]
      attrs[:modified_at] = modified_at
    end

    if created_at = record_node["createdAt"]
      attrs[:created_at] = created_at
    end

    if work_node = record_node["work"]
      attrs[:work] = WorkEntity.from_node(work_node)
    end

    itemable_node = record_node["itemable"]
    if itemable_type && itemable_node
      attrs[:itemable] = case itemable_type
      when "EPISODE_RECORD"
        EpisodeRecordEntity.from_node(itemable_node)
      when "WORK_RECORD"
        WorkRecordEntity.from_node(itemable_node)
      end
    end

    new attrs
  end
end
