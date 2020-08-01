# frozen_string_literal: true

class ActivityGroupEntity < ApplicationEntity
  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :itemable_type, Types::ActivityResourceKinds
  attribute? :single, Types::Bool
  attribute? :activities_count, Types::Integer
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :itemables, Types::Array.of(EpisodeRecordEntity | StatusEntity | AnimeRecordEntity)
  attribute? :activities_page_info, PageInfoEntity

  def self.from_nodes(activity_group_nodes)
    activity_group_nodes.map do |activity_group_node|
      from_node(activity_group_node)
    end
  end

  def self.from_node(activity_group_node)
    attrs = {}

    if id = activity_group_node["id"]
      attrs[:id] = id
    end

    if itemable_type = activity_group_node["itemableType"]
      attrs[:itemable_type] = itemable_type.downcase
    end

    if single = activity_group_node["single"]
      attrs[:single] = single
    end

    if activities_count = activity_group_node["activitiesCount"]
      attrs[:activities_count] = activities_count
    end

    if created_at = activity_group_node["createdAt"]
      attrs[:created_at] = created_at
    end

    if user_node = activity_group_node["user"]
      attrs[:user] = UserEntity.from_node(user_node)
    end

    activity_nodes = activity_group_node.dig("activities", "nodes")
    attrs[:itemables] = user_node && activity_nodes ? ActivityEntity.from_nodes(activity_nodes, user_node: user_node).map(&:itemable) : []

    if page_info_node = activity_group_node.dig("activities", "pageInfo")
      attrs[:activities_page_info] = PageInfoEntity.from_node(page_info_node)
    end

    new attrs
  end

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
