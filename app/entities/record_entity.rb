# frozen_string_literal: true

class RecordEntity < ApplicationEntity
  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :comment, Types::String.optional
  attribute? :comment_html, Types::String.optional
  attribute? :likes_count, Types::Integer
  attribute? :modified_at, Types::Params::Time.optional
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :trackable, AnimeEntity | EpisodeEntity
  attribute? :recordable, AnimeRecordEntity | EpisodeRecordEntity

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

    if trackable_node = node["trackable"]
      attrs[:trackable] = case trackable_node["__typename"]
      when "Anime"
        AnimeEntity.from_node(trackable_node)
      when "Episode"
        EpisodeEntity.from_node(trackable_node)
      end
    end

    if recordable_node = node["recordable"]
      attrs[:recordable] = case recordable_node["__typename"]
      when "AnimeRecord"
        AnimeRecordEntity.from_node(recordable_node)
      when "EpisodeRecord"
        EpisodeRecordEntity.from_node(recordable_node)
      end
    end

    new attrs
  end

  def episode_record?
    recordable.is_a?(EpisodeRecordEntity)
  end

  def anime_record?
    recordable.is_a?(AnimeRecordEntity)
  end
end
