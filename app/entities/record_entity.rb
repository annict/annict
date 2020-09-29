# frozen_string_literal: true

class RecordEntity < ApplicationEntity
  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :complementable_type, Types::RecordComplementableTypes
  attribute? :comment, Types::String.optional
  attribute? :comment_html, Types::String.optional
  attribute? :likes_count, Types::Integer
  attribute? :modified_at, Types::Params::Time.optional
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :trackable, AnimeEntity | EpisodeEntity
  attribute? :complementable, AnimeRecordEntity | EpisodeRecordEntity

  def self.from_nodes(nodes)
    nodes.map do |node|
      from_node(node)
    end
  end

  def self.from_node(node)
    attrs = {}

    if database_id = node["databaseId"]
      attrs[:database_id] = database_id
    end

    if complementable_type = node["complementableType"]
      attrs[:complementable_type] = complementable_type.downcase
    end

    if comment = node["comment"]
      attrs[:comment] = comment
    end

    if comment_html = node["commentHtml"]
      attrs[:comment_html] = comment_html
    end

    if likes_count = node["likesCount"]
      attrs[:likes_count] = likes_count
    end

    if modified_at = node["modifiedAt"]
      attrs[:modified_at] = modified_at
    end

    if created_at = node["createdAt"]
      attrs[:created_at] = created_at
    end

    if user_node = node["user"]
      attrs[:user] = UserEntity.from_node(user_node)
    end

    trackable_node = node["trackable"]
    if complementable_type && trackable_node
      attrs[:trackable] = case complementable_type
      when "ANIME_RECORD"
        AnimeEntity.from_node(trackable_node)
      when "EPISODE_RECORD"
        EpisodeEntity.from_node(trackable_node)
      end
    end

    complementable_node = node["complementable"]
    if complementable_type && complementable_node
      attrs[:complementable] = case complementable_type
      when "ANIME_RECORD"
        AnimeRecordEntity.from_node(complementable_node)
      when "EPISODE_RECORD"
        EpisodeRecordEntity.from_node(complementable_node)
      end
    end

    new attrs
  end
end
