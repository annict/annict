# frozen_string_literal: true

class WorkObject < ApplicationObject
  attribute :id, ObjectTypes::Strict::String
  attribute :annict_id, ObjectTypes::Strict::Integer
  attribute :title, ObjectTypes::Strict::String
  attribute :watchers_count, ObjectTypes::Strict::Integer
  attribute :copyright, ObjectTypes::Strict::String
  attribute :satisfaction_rate, ObjectTypes::Strict::Float
  attribute :ratings_count, ObjectTypes::Strict::Integer
  attribute :title_kana, ObjectTypes::Strict::String
  attribute :official_site_url, ObjectTypes::Strict::String
  attribute :twitter_username, ObjectTypes::Strict::String
  attribute :wikipedia_url, ObjectTypes::Strict::String
  attribute :is_no_episodes, ObjectTypes::Strict::Bool
  attribute :synopsis, ObjectTypes::Strict::String
  attribute :synopsis_en, ObjectTypes::Strict::String
  attribute :synopsis_source, ObjectTypes::Strict::String

  attribute :image, WorkImageObject
  attribute :trailers, ObjectTypes::Strict::Array.of(TrailerObject)
  attribute :casts, ObjectTypes::Strict::Array.of(CastObject)
  attribute :staffs, ObjectTypes::Strict::Array.of(StaffObject)
  attribute :episodes, ObjectTypes::Strict::Array.of(EpisodeObject)
end
