# frozen_string_literal: true

class WorkStruct < ApplicationStruct
  attribute :id, StructTypes::Strict::String.optional.default(nil)
  attribute :annict_id, StructTypes::Strict::Integer.optional.default(nil)
  attribute :title, StructTypes::Strict::String.optional.default(nil)
  attribute :watchers_count, StructTypes::Strict::Integer.optional.default(nil)
  attribute :copyright, StructTypes::Strict::String.optional.default(nil)
  attribute :satisfaction_rate, StructTypes::Strict::Float.optional.default(nil)
  attribute :ratings_count, StructTypes::Strict::Integer.optional.default(nil)
  attribute :title_kana, StructTypes::Strict::String.optional.default(nil)
  attribute :official_site_url, StructTypes::Strict::String.optional.default(nil)
  attribute :twitter_username, StructTypes::Strict::String.optional.default(nil)
  attribute :wikipedia_url, StructTypes::Strict::String.optional.default(nil)
  attribute :is_no_episodes, StructTypes::Strict::Bool.optional.default(nil)
  attribute :synopsis, StructTypes::Strict::String.optional.default(nil)
  attribute :synopsis_en, StructTypes::Strict::String.optional.default(nil)
  attribute :synopsis_source, StructTypes::Strict::String.optional.default(nil)
  attribute :viewer_status_state, StructTypes::Strict::String.optional.default(nil)

  attribute :image, WorkImageStruct.optional.default(nil)
  attribute :trailers, StructTypes::Strict::Array.of(TrailerStruct).optional.default(nil)
  attribute :casts, StructTypes::Strict::Array.of(CastStruct).optional.default(nil)
  attribute :staffs, StructTypes::Strict::Array.of(StaffStruct).optional.default(nil)
  attribute :episodes, StructTypes::Strict::Array.of(EpisodeStruct).optional.default(nil)
end
