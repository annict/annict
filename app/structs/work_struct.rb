# frozen_string_literal: true

class WorkStruct < ApplicationStruct
  attribute :id, StructTypes::Strict::String
  attribute :annict_id, StructTypes::Strict::Integer
  attribute :title, StructTypes::Strict::String
  attribute :watchers_count, StructTypes::Strict::Integer
  attribute :copyright, StructTypes::Strict::String
  attribute :satisfaction_rate, StructTypes::Strict::Float
  attribute :ratings_count, StructTypes::Strict::Integer
  attribute :title_kana, StructTypes::Strict::String
  attribute :official_site_url, StructTypes::Strict::String
  attribute :twitter_username, StructTypes::Strict::String
  attribute :wikipedia_url, StructTypes::Strict::String
  attribute :is_no_episodes, StructTypes::Strict::Bool
  attribute :synopsis, StructTypes::Strict::String
  attribute :synopsis_en, StructTypes::Strict::String
  attribute :synopsis_source, StructTypes::Strict::String

  attribute :image, WorkImageStruct
  attribute :trailers, StructTypes::Strict::Array.of(TrailerStruct)
  attribute :casts, StructTypes::Strict::Array.of(CastStruct)
  attribute :staffs, StructTypes::Strict::Array.of(StaffStruct)
  attribute :episodes, StructTypes::Strict::Array.of(EpisodeStruct)
end
