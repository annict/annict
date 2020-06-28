# frozen_string_literal: true

class ActivityEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :itemable_type, Types::ActivityResourceKinds
  attribute? :itemable, EpisodeRecordEntity | StatusEntity | WorkRecordEntity

  def self.from_nodes(activity_nodes, user_node: nil)
    activity_nodes.map do |activity_node|
      from_node(activity_node, user_node: user_node)
    end
  end

  def self.from_node(activity_node, user_node: nil)
    attrs = {}

    if database_id = activity_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if itemable_type = activity_node["itemableType"]
      attrs[:itemable_type] = itemable_type.downcase
    end

    if itemable_node = activity_node["itemable"]
      attrs[:itemable] = case itemable_type
      when "EPISODE_RECORD"
        EpisodeRecordEntity.from_node(itemable_node, user_node: user_node)
      when "STATUS"
        StatusEntity.from_node(itemable_node, user_node: user_node)
      when "WORK_RECORD"
        WorkRecordEntity.from_node(itemable_node, user_node: user_node)
      end
    end

    new attrs
  end
end
