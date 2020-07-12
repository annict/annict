# frozen_string_literal: true

class StatusEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :kind, Types::StatusKinds
  attribute? :likes_count, Types::Integer
  attribute? :user, UserEntity
  attribute? :work, WorkEntity

  def self.from_node(status_node, user_node: nil)
    attrs = {}

    if database_id = status_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if kind = status_node["kind"]
      attrs[:kind] = kind.downcase
    end

    if likes_count = status_node["likesCount"]
      attrs[:likes_count] = likes_count
    end

    if work_node = status_node["work"]
      attrs[:work] = WorkEntity.from_node(work_node)
    end

    if user_node = status_node["user"] || user_node
      attrs[:user] = UserEntity.from_node(user_node)
    end

    new attrs
  end
end
